package sync

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/someone235/katnip/server/database"
	"github.com/someone235/katnip/server/serializer"

	"github.com/someone235/katnip/server/dbaccess"
	"github.com/someone235/katnip/server/dbmodels"
)

func insertTransactionInputs(dbTx *database.TxContext, transactionHashesToTxsWithMetadata map[string]*txWithMetadata) error {
	onEnd := logger.LogAndMeasureExecutionTime(log, "insertTransactionInputs")
	defer onEnd()

	outpointsSet := make(map[dbaccess.Outpoint]struct{})
	newNonCoinbaseTransactions := make(map[string]*txWithMetadata)
	inputsCount := 0
	for txHash, transaction := range transactionHashesToTxsWithMetadata {
		if !transaction.isNew {
			continue
		}
		isCoinbase, err := isTransactionCoinbase(transaction.verboseTx)
		if err != nil {
			return err
		}
		if isCoinbase {
			continue
		}

		newNonCoinbaseTransactions[txHash] = transaction
		inputsCount += len(transaction.verboseTx.TransactionVerboseInputs)
		for i := range transaction.verboseTx.TransactionVerboseInputs {
			txIn := transaction.verboseTx.TransactionVerboseInputs[i]
			outpoint := dbaccess.Outpoint{
				TransactionID: txIn.TxID,
				Index:         txIn.OutputIndex,
			}
			outpointsSet[outpoint] = struct{}{}
		}
	}

	if inputsCount == 0 {
		return nil
	}
	outpoints := make([]*dbaccess.Outpoint, len(outpointsSet))
	i := 0
	for outpoint := range outpointsSet {
		outpointCopy := outpoint // since outpoint is a value type - copy it, othewise it would be overwritten
		outpoints[i] = &outpointCopy
		i++
	}

	dbPreviousTransactionsOutputs, err := dbaccess.TransactionOutputsByOutpoints(dbTx, outpoints)
	if err != nil {
		return err
	}

	if len(dbPreviousTransactionsOutputs) != len(outpoints) {
		log.Debugf("couldn't fetch all of the requested outpoints")
	}

	outpointsToIDs := make(map[dbaccess.Outpoint]uint64)
	for _, dbTransactionOutput := range dbPreviousTransactionsOutputs {
		outpointsToIDs[dbaccess.Outpoint{
			TransactionID: dbTransactionOutput.Transaction.TransactionID,
			Index:         dbTransactionOutput.Index,
		}] = dbTransactionOutput.ID
	}

	inputsToAdd := make([]interface{}, inputsCount)
	inputIndex := 0
	for _, transaction := range newNonCoinbaseTransactions {
		for i, txIn := range transaction.verboseTx.TransactionVerboseInputs {
			scriptSig, err := hex.DecodeString(txIn.ScriptSig.Hex)
			if err != nil {
				return nil
			}
			dbTransactionInput := &dbmodels.TransactionInput{
				TransactionID:                  transaction.id,
				PreviousTransactionOutputIndex: txIn.OutputIndex,
				PreviousTransactionID:          txIn.TxID,
				Index:                          uint32(i),
				SignatureScript:                scriptSig,
				Sequence:                       serializer.Uint64ToBytes(txIn.Sequence),
			}

			prevOutputID, ok := outpointsToIDs[dbaccess.Outpoint{
				TransactionID: txIn.TxID,
				Index:         txIn.OutputIndex,
			}]
			if ok && prevOutputID != 0 {
				dbTransactionInput.PreviousTransactionOutputID = prevOutputID
			}

			inputsToAdd[inputIndex] = dbTransactionInput
			inputIndex++
		}
	}
	return dbaccess.BulkInsert(dbTx, inputsToAdd)
}

func isTransactionCoinbase(transaction *appmessage.TransactionVerboseData) (bool, error) {
	subnetwork, err := subnetworks.FromString(transaction.SubnetworkID)
	if err != nil {
		return false, err
	}
	return subnetwork.Equal(&subnetworks.SubnetworkIDCoinbase), nil
}
