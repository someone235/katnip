package sync

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/someone235/katnip/server/database"

	"github.com/pkg/errors"
	"github.com/someone235/katnip/server/dbaccess"
	"github.com/someone235/katnip/server/dbmodels"
)

func insertTransactionOutputs(dbTx *database.TxContext, transactionHashesToTxsWithMetadata map[string]*txWithMetadata) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "insertTransactionOutputs")
	defer onEnd()

	addressesToAddressIDs, err := insertAddresses(dbTx, transactionHashesToTxsWithMetadata)
	if err != nil {
		return err
	}

	outputsToAdd := make([]interface{}, 0)
	for _, transaction := range transactionHashesToTxsWithMetadata {
		if !transaction.isNew {
			continue
		}
		for i, txOut := range transaction.tx.Outputs {
			scriptPubKey, err := hex.DecodeString(txOut.ScriptPublicKey.Script)
			if err != nil {
				return errors.WithStack(err)
			}
			var addressID *uint64
			if txOut.VerboseData.ScriptPublicKeyAddress != "" {
				addressIDValue := addressesToAddressIDs[txOut.VerboseData.ScriptPublicKeyAddress]
				addressID = &addressIDValue
			}
			outputsToAdd = append(outputsToAdd, &dbmodels.TransactionOutput{
				TransactionID: transaction.id,
				Index:         uint32(i),
				Value:         txOut.Amount,
				IsSpent:       false, // This must be false for updateSelectedParentChain to work properly
				ScriptPubKey:  scriptPubKey,
				AddressID:     addressID,
			})
		}
	}

	return dbaccess.BulkInsert(dbTx, outputsToAdd)
}
