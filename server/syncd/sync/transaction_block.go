package sync

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/someone235/katnip/server/database"
	"github.com/someone235/katnip/server/dbaccess"
	"github.com/someone235/katnip/server/dbmodels"

	"github.com/pkg/errors"
)

func insertTransactionBlocks(dbTx *database.TxContext, blocks []*appmessage.BlockVerboseData,
	blockHashesToIDs map[string]uint64, transactionHashesToTxsWithMetadata map[string]*txWithMetadata) error {

	transactionBlocksToAdd := make([]interface{}, 0)
	for _, block := range blocks {
		blockID, ok := blockHashesToIDs[block.Hash]
		if !ok {
			return errors.Errorf("couldn't find block ID for block %s", block.Hash)
		}
		for i, tx := range block.TransactionVerboseData {
			transactionBlocksToAdd = append(transactionBlocksToAdd, &dbmodels.TransactionBlock{
				TransactionID: transactionHashesToTxsWithMetadata[tx.Hash].id,
				BlockID:       blockID,
				Index:         uint32(i),
			})
		}
	}
	return dbaccess.BulkInsert(dbTx, transactionBlocksToAdd)
}
