package sync

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kasparov/database"

	"github.com/kaspanet/kasparov/dbaccess"
	"github.com/kaspanet/kasparov/kaspadrpc"
	"github.com/kaspanet/kasparov/kasparovsyncd/mqtt"
)

// StartSync keeps the node and the database in sync. On start, it downloads
// all data that's missing from the database, and once it's done it keeps
// sync with the node via notifications.
func StartSync(doneChan chan struct{}) error {
	client, err := kaspadrpc.GetClient()
	if err != nil {
		return err
	}

	// Mass download missing data
	err = fetchInitialData(client)
	if err != nil {
		return err
	}

	// Keep the node and the database in sync
	return sync(client, doneChan)
}

// fetchInitialData downloads all data that's currently missing from
// the database.
func fetchInitialData(client *kaspadrpc.Client) error {
	log.Infof("Syncing past blocks")
	err := syncBlocks(client)
	if err != nil {
		return err
	}
	log.Infof("Finished syncing past data")
	return nil
}

// sync keeps the database in sync with the node via notifications
func sync(client *kaspadrpc.Client, doneChan chan struct{}) error {
	// Handle client notifications until we're told to stop
	for {
		select {
		case blockAdded := <-client.OnBlockAdded:
			err := handleBlockAddedMsg(client, blockAdded)
			if err != nil {
				return err
			}
		case <-doneChan:
			log.Infof("StartSync stopped")
			return nil
		}
	}
}

// syncBlocks attempts to download all DAG blocks starting with
// the bluest block, and then inserts them into the database.
func syncBlocks(client *kaspadrpc.Client) error {
	// Start syncing from the bluest block hash. We use blue score to
	// simulate the "last" block we have because blue-block order is
	// the order that the node uses in the various JSONRPC calls.
	startBlock, err := dbaccess.BluestBlock(database.NoTx())
	if err != nil {
		return err
	}
	var startHash string
	if startBlock != nil {
		startHash = startBlock.BlockHash
	}

	for {
		if startHash != "" {
			log.Debugf("Calling getBlocks with start hash %s", startHash)
		} else {
			log.Debugf("Calling getBlocks with no start hash")
		}
		blocksResult, err := client.GetBlocks(startHash, true, true)
		if err != nil {
			return err
		}
		if startHash != "" && len(blocksResult.BlockHashes) == 1 {
			break
		}
		log.Debugf("Got %d blocks", len(blocksResult.BlockHashes))

		startHash = blocksResult.BlockHashes[len(blocksResult.BlockHashes)-1]
		err = addBlocks(client, blocksResult)
		if err != nil {
			return err
		}
	}

	return nil
}

// fetchBlock downloads the serialized block and raw block data of
// the block with hash blockHash.
func fetchBlock(client *kaspadrpc.Client, blockHash *externalapi.DomainHash) (
	*appmessage.BlockVerboseData, error) {
	log.Debugf("Getting block %s from the RPC server", blockHash)
	blockHexResponse, err := client.GetBlock(blockHash.String(), true)
	if err != nil {
		return nil, err
	}
	return blockHexResponse.BlockVerboseData, nil
}

func handleBlockAddedMsg(client *kaspadrpc.Client, blockAdded *appmessage.BlockAddedNotificationMessage) error {
	blockHash := blockAdded.Block.Header.BlockHash()
	blockExists, err := dbaccess.DoesBlockExist(database.NoTx(), blockHash.String())
	if err != nil {
		return err
	}
	if blockExists {
		return nil
	}

	dbTx, err := database.NewTx()
	if err != nil {
		return err
	}

	defer dbTx.RollbackUnlessCommitted()

	addedBlockHashes, err := fetchAndAddBlock(client, dbTx, blockHash)
	if err != nil {
		return err
	}

	err = dbTx.Commit()
	if err != nil {
		return err
	}

	for _, hash := range addedBlockHashes {
		err := mqtt.PublishBlockAddedNotifications(hash)
		if err != nil {
			return err
		}
	}
	return nil
}

func fetchAndAddBlock(client *kaspadrpc.Client, dbTx *database.TxContext,
	blockHash *externalapi.DomainHash) (addedBlockHashes []string, err error) {

	block, err := fetchBlock(client, blockHash)
	if err != nil {
		return nil, err
	}

	missingAncestors, err := fetchMissingAncestors(client, dbTx, block, nil)
	if err != nil {
		return nil, err
	}

	blocks := append([]*appmessage.BlockVerboseData{block}, missingAncestors...)
	err = bulkInsertBlocksData(client, dbTx, blocks)
	if err != nil {
		return nil, err
	}

	addedBlockHashes = make([]string, len(blocks))
	for i, block := range blocks {
		addedBlockHashes[i] = block.Hash
	}

	return addedBlockHashes, nil
}

func fetchMissingAncestors(client *kaspadrpc.Client, dbTx *database.TxContext, block *appmessage.BlockVerboseData,
	blockExistingInMemory map[string]*appmessage.BlockVerboseData) ([]*appmessage.BlockVerboseData, error) {

	pendingBlocks := []*appmessage.BlockVerboseData{block}
	missingAncestors := make([]*appmessage.BlockVerboseData, 0)
	missingAncestorsSet := make(map[string]struct{})
	for len(pendingBlocks) > 0 {
		var currentBlock *appmessage.BlockVerboseData
		currentBlock, pendingBlocks = pendingBlocks[0], pendingBlocks[1:]
		missingParentHashes, err := missingBlockHashes(dbTx, currentBlock.ParentHashes, blockExistingInMemory)
		if err != nil {
			return nil, err
		}
		blocksToPrependToPending := make([]*appmessage.BlockVerboseData, 0, len(missingParentHashes))
		for _, missingHash := range missingParentHashes {
			if _, ok := missingAncestorsSet[missingHash]; ok {
				continue
			}
			hash, err := externalapi.NewDomainHashFromString(missingHash)
			if err != nil {
				return nil, err
			}
			block, err := fetchBlock(client, hash)
			if err != nil {
				return nil, err
			}
			blocksToPrependToPending = append(blocksToPrependToPending, block)
		}
		if len(blocksToPrependToPending) == 0 {
			if currentBlock != block {
				missingAncestorsSet[currentBlock.Hash] = struct{}{}
				missingAncestors = append(missingAncestors, currentBlock)
			}
			continue
		}
		log.Debugf("Found %d missing parents for block %s and fetched them", len(blocksToPrependToPending), currentBlock.Hash)
		blocksToPrependToPending = append(blocksToPrependToPending, currentBlock)
		pendingBlocks = append(blocksToPrependToPending, pendingBlocks...)
	}
	return missingAncestors, nil
}

// missingBlockHashes takes a slice of block hashes and returns
// a slice that contains all the block hashes that do not exist
// in the database or in the given blocksExistingInMemory map.
func missingBlockHashes(dbTx *database.TxContext, blockHashes []string,
	blocksExistingInMemory map[string]*appmessage.BlockVerboseData) ([]string, error) {

	// filter out all the hashes that exist in blocksExistingInMemory
	hashesNotInMemory := make([]string, 0)
	for _, hash := range blockHashes {
		if _, ok := blocksExistingInMemory[hash]; !ok {
			hashesNotInMemory = append(hashesNotInMemory, hash)
		}
	}

	// Check which of the hashes in hashesNotInMemory do
	// not exist in the database.
	dbBlocks, err := dbaccess.BlocksByHashes(dbTx, hashesNotInMemory)
	if err != nil {
		return nil, err
	}
	if len(hashesNotInMemory) != len(dbBlocks) {
		// Some hashes are missing. Collect and return them
		var missingHashes []string
	outerLoop:
		for _, hash := range hashesNotInMemory {
			for _, dbBlock := range dbBlocks {
				if dbBlock.BlockHash == hash {
					continue outerLoop
				}
			}
			missingHashes = append(missingHashes, hash)
		}
		return missingHashes, nil
	}

	return nil, nil
}

// addBlocks inserts data in the given verbose blocks
// into the database.
func addBlocks(client *kaspadrpc.Client, getBlocksResponse *appmessage.GetBlocksResponseMessage) error {
	dbTx, err := database.NewTx()
	if err != nil {
		return err
	}
	defer dbTx.RollbackUnlessCommitted()

	nonExistingBlocks, err := getNonExistingBlocks(dbTx, getBlocksResponse)
	if err != nil {
		return err
	}

	err = bulkInsertBlocksData(client, dbTx, nonExistingBlocks)
	if err != nil {
		return err
	}
	return dbTx.Commit()
}

// bulkInsertBlocksData inserts the given blocks and their data (transactions
// and new subnetworks data) to the database in chunks.
func bulkInsertBlocksData(client *kaspadrpc.Client, dbTx *database.TxContext, verboseBlocks []*appmessage.BlockVerboseData) error {
	subnetworkIDToID, err := insertSubnetworks(client, dbTx, verboseBlocks)
	if err != nil {
		return err
	}

	transactionHashesToTxsWithMetadata, err := insertTransactions(dbTx, verboseBlocks, subnetworkIDToID)
	if err != nil {
		return err
	}

	err = insertTransactionOutputs(dbTx, transactionHashesToTxsWithMetadata)
	if err != nil {
		return err
	}

	err = insertTransactionInputs(dbTx, transactionHashesToTxsWithMetadata)
	if err != nil {
		return err
	}

	err = insertBlocks(dbTx, verboseBlocks)
	if err != nil {
		return err
	}

	blockHashesToIDs, err := getBlocksWithTheirParentIDs(dbTx, verboseBlocks)
	if err != nil {
		return err
	}

	err = insertBlockParents(dbTx, verboseBlocks, blockHashesToIDs)
	if err != nil {
		return err
	}

	err = insertTransactionBlocks(dbTx, verboseBlocks, blockHashesToIDs, transactionHashesToTxsWithMetadata)
	if err != nil {
		return err
	}

	log.Infof("Added %d blocks", len(verboseBlocks))
	return nil
}
