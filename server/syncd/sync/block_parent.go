package sync

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/someone235/katnip/server/database"
	"github.com/someone235/katnip/server/dbaccess"
	"github.com/someone235/katnip/server/dbmodels"

	"github.com/pkg/errors"
)

func insertBlockParents(dbTx *database.TxContext, blocks []*appmessage.RPCBlock, blockHashesToIDs map[string]uint64) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "insertBlockParents")
	defer onEnd()

	parentsToAdd := make([]interface{}, 0)
	for _, block := range blocks {
		dbBlockParents, err := dbParentBlocksFromRPCBlock(blockHashesToIDs, block)
		if err != nil {
			return err
		}
		for _, dbBlockParent := range dbBlockParents {
			parentsToAdd = append(parentsToAdd, dbBlockParent)
		}
	}
	err := dbaccess.BulkInsert(dbTx, parentsToAdd)
	if err != nil {
		return err
	}
	return nil
}

func dbParentBlocksFromRPCBlock(blockHashesToIDs map[string]uint64, block *appmessage.RPCBlock) ([]*dbmodels.ParentBlock, error) {
	// Exit early if this is the genesis block
	if len(block.Header.ParentHashes) == 0 {
		return nil, nil
	}

	blockID, ok := blockHashesToIDs[block.VerboseData.Hash]
	if !ok {
		return nil, errors.Errorf("couldn't find block ID for block %s", block.VerboseData.Hash)
	}
	dbParentBlocks := make([]*dbmodels.ParentBlock, len(block.Header.ParentHashes))
	for i, parentHash := range block.Header.ParentHashes {
		parentID, ok := blockHashesToIDs[parentHash]
		if !ok {
			return nil, errors.Errorf("missing parent %s for block %s", parentHash, block.VerboseData.Hash)
		}
		dbParentBlocks[i] = &dbmodels.ParentBlock{
			BlockID:       blockID,
			ParentBlockID: parentID,
		}
	}
	return dbParentBlocks, nil
}
