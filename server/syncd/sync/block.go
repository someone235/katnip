package sync

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/kaspanet/kaspad/util/mstime"
	"github.com/someone235/katnip/server/database"
	"github.com/someone235/katnip/server/serializer"

	"github.com/pkg/errors"
	"github.com/someone235/katnip/server/dbaccess"
	"github.com/someone235/katnip/server/dbmodels"
)

func insertBlocks(dbTx *database.TxContext, blocks []*appmessage.RPCBlock) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "insertBlocks")
	defer onEnd()

	blocksToAdd := make([]interface{}, len(blocks))
	for i, block := range blocks {
		var err error
		blocksToAdd[i], err = dbBlockFromRPCBlock(block)
		if err != nil {
			return err
		}
	}
	return dbaccess.BulkInsert(dbTx, blocksToAdd)
}

func getBlocksWithTheirParentIDs(dbTx *database.TxContext, blocks []*appmessage.RPCBlock) (map[string]uint64, error) {
	onEnd := logger.LogAndMeasureExecutionTime(log, "getBlocksWithTheirParentIDs")
	defer onEnd()

	blockSet := make(map[string]struct{})
	for _, block := range blocks {
		blockSet[block.VerboseData.Hash] = struct{}{}
		for _, parentHash := range block.Header.ParentHashes {
			blockSet[parentHash] = struct{}{}
		}
	}

	blockHashes := stringsSetToSlice(blockSet)

	dbBlocks, err := dbaccess.BlocksByHashes(dbTx, blockHashes)
	if err != nil {
		return nil, err
	}

	if len(dbBlocks) != len(blockHashes) {
		for _, hash := range blockHashes {
			block, err := dbaccess.BlockByHash(dbTx, hash)
			if err != nil {
				return nil, err
			}

			if block == nil {
				return nil, errors.Errorf("couldn't retrieve block %s", hash)
			}
		}
		return nil, errors.Errorf("couldn't find the missing block")
	}

	blockHashesToIDs := make(map[string]uint64)
	for _, dbBlock := range dbBlocks {
		blockHashesToIDs[dbBlock.BlockHash] = dbBlock.ID
	}
	return blockHashesToIDs, nil
}

func getNonExistingBlocks(dbTx *database.TxContext, getBlocksResponse *appmessage.GetBlocksResponseMessage) (
	[]*appmessage.RPCBlock, error) {

	existingBlockHashes, err := dbaccess.ExistingHashes(dbTx, getBlocksResponse.BlockHashes)
	if err != nil {
		return nil, err
	}

	existingBlockHashesSet := make(map[string]struct{}, len(existingBlockHashes))
	for _, hash := range existingBlockHashes {
		existingBlockHashesSet[hash] = struct{}{}
	}

	nonExistingBlocks := make([]*appmessage.RPCBlock, 0, len(getBlocksResponse.Blocks))
	for _, block := range getBlocksResponse.Blocks {
		if _, exists := existingBlockHashesSet[block.VerboseData.Hash]; exists {
			continue
		}

		nonExistingBlocks = append(nonExistingBlocks, block)
	}

	return nonExistingBlocks, nil
}

func dbBlockFromRPCBlock(block *appmessage.RPCBlock) (*dbmodels.Block, error) {
	dbBlock := dbmodels.Block{
		BlockHash:            block.VerboseData.Hash,
		Version:              uint16(block.Header.Version),
		HashMerkleRoot:       block.Header.HashMerkleRoot,
		AcceptedIDMerkleRoot: block.Header.AcceptedIDMerkleRoot,
		UTXOCommitment:       block.Header.UTXOCommitment,
		Timestamp:            mstime.UnixMilliseconds(block.Header.Timestamp).ToNativeTime(),
		Bits:                 block.Header.Bits,
		Nonce:                serializer.Uint64ToBytes(block.Header.Nonce),
		BlueScore:            block.VerboseData.BlueScore,
		IsChainBlock:         false, // This must be false for updateSelectedParentChain to work properly
		TransactionCount:     uint16(len(block.Transactions)),
		Difficulty:           block.VerboseData.Difficulty,
	}

	// Set genesis block as the initial chain block
	if len(block.Header.ParentHashes) == 0 {
		dbBlock.IsChainBlock = true
	}
	return &dbBlock, nil
}
