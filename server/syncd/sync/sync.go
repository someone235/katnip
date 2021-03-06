package sync

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/someone235/katnip/server/database"

	"github.com/someone235/katnip/server/dbaccess"
	"github.com/someone235/katnip/server/kaspadrpc"
	"github.com/someone235/katnip/server/syncd/mqtt"
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

		err = addBlocks(client, blocksResult)
		if err != nil {
			return err
		}

		startHash = blocksResult.BlockHashes[len(blocksResult.BlockHashes)-1]
	}

	return nil
}

// fetchBlock downloads the serialized block and raw block data of
// the block with hash blockHash.
func fetchBlock(client *kaspadrpc.Client, blockHash string) (
	*appmessage.RPCBlock, error) {
	log.Debugf("Getting block %s from the RPC server", blockHash)
	blockResponse, err := client.GetBlock(blockHash, true)
	if err != nil {
		return nil, err
	}
	return blockResponse.Block, nil
}

func handleBlockAddedMsg(client *kaspadrpc.Client, blockAdded *appmessage.BlockAddedNotificationMessage) error {
	blockHash := blockAdded.Block.VerboseData.Hash
	blockExists, err := dbaccess.DoesBlockExist(database.NoTx(), blockHash)
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
	blockHash string) (addedBlockHashes []string, err error) {

	block, err := fetchBlock(client, blockHash)
	if err != nil {
		return nil, err
	}

	missingAncestors, err := fetchMissingAncestors(client, dbTx, block, nil)
	if err != nil {
		return nil, err
	}

	blocks := append([]*appmessage.RPCBlock{block}, missingAncestors...)
	err = bulkInsertBlocksData(client, dbTx, blocks)
	if err != nil {
		return nil, err
	}

	addedBlockHashes = make([]string, len(blocks))
	for i, block := range blocks {
		addedBlockHashes[i] = block.VerboseData.Hash
	}

	return addedBlockHashes, nil
}

func fetchMissingAncestors(client *kaspadrpc.Client, dbTx *database.TxContext, block *appmessage.RPCBlock,
	blockExistingInMemory map[string]*appmessage.RPCBlock) ([]*appmessage.RPCBlock, error) {

	pendingBlocks := []*appmessage.RPCBlock{block}
	missingAncestors := make([]*appmessage.RPCBlock, 0)
	missingAncestorsSet := make(map[string]struct{})
	for len(pendingBlocks) > 0 {
		var currentBlock *appmessage.RPCBlock
		currentBlock, pendingBlocks = pendingBlocks[0], pendingBlocks[1:]
		missingParentHashes, err := missingBlockHashes(dbTx, currentBlock.Header.ParentHashes, blockExistingInMemory)
		if err != nil {
			return nil, err
		}
		blocksToPrependToPending := make([]*appmessage.RPCBlock, 0, len(missingParentHashes))
		for _, missingHash := range missingParentHashes {
			if _, ok := missingAncestorsSet[missingHash]; ok {
				continue
			}
			block, err := fetchBlock(client, missingHash)
			if err != nil {
				return nil, err
			}
			blocksToPrependToPending = append(blocksToPrependToPending, block)
		}
		if len(blocksToPrependToPending) == 0 {
			if currentBlock != block {
				missingAncestorsSet[currentBlock.VerboseData.Hash] = struct{}{}
				missingAncestors = append(missingAncestors, currentBlock)
			}
			continue
		}
		log.Debugf("Found %d missing parents for block %s and fetched them",
			len(blocksToPrependToPending), currentBlock.VerboseData.Hash)
		blocksToPrependToPending = append(blocksToPrependToPending, currentBlock)
		pendingBlocks = append(blocksToPrependToPending, pendingBlocks...)
	}
	return missingAncestors, nil
}

// missingBlockHashes takes a slice of block hashes and returns
// a slice that contains all the block hashes that do not exist
// in the database or in the given blocksExistingInMemory map.
func missingBlockHashes(dbTx *database.TxContext, blockHashes []string,
	blocksExistingInMemory map[string]*appmessage.RPCBlock) ([]string, error) {

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
func bulkInsertBlocksData(client *kaspadrpc.Client, dbTx *database.TxContext, blocks []*appmessage.RPCBlock) error {
	subnetworkIDToID, err := insertSubnetworks(client, dbTx, blocks)
	if err != nil {
		return err
	}

	transactionHashesToTxsWithMetadata, err := insertTransactions(dbTx, blocks, subnetworkIDToID)
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

	err = insertBlocks(dbTx, blocks)
	if err != nil {
		return err
	}

	blockHashesToIDs, err := getBlocksWithTheirParentIDs(dbTx, blocks)
	if err != nil {
		return err
	}

	err = insertBlockParents(dbTx, blocks, blockHashesToIDs)
	if err != nil {
		return err
	}

	err = insertTransactionBlocks(dbTx, blocks, blockHashesToIDs, transactionHashesToTxsWithMetadata)
	if err != nil {
		return err
	}

	log.Infof("Added %d blocks", len(blocks))
	return nil
}
