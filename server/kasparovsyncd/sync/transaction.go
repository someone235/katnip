package sync

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/kaspanet/kaspad/domain/consensus/utils/subnetworks"
	"github.com/kaspanet/kasparov/database"
	"github.com/kaspanet/kasparov/serializer"

	"github.com/kaspanet/kasparov/dbaccess"
	"github.com/kaspanet/kasparov/dbmodels"

	"github.com/pkg/errors"
)

type txWithMetadata struct {
	verboseTx *appmessage.TransactionVerboseData
	id        uint64
	isNew     bool
}

func transactionHashesToTxsWithMetadataToTransactionHashes(transactionHashesToTxsWithMetadata map[string]*txWithMetadata) []string {
	hashes := make([]string, len(transactionHashesToTxsWithMetadata))
	i := 0
	for hash := range transactionHashesToTxsWithMetadata {
		hashes[i] = hash
		i++
	}
	return hashes
}

func insertTransactions(dbTx *database.TxContext, blocks []*appmessage.BlockVerboseData, subnetworkIDsToIDs map[string]uint64) (
	map[string]*txWithMetadata, error) {

	transactionHashesToTxsWithMetadata := make(map[string]*txWithMetadata)
	for _, block := range blocks {
		// We do not directly iterate over block.Verbose.RawTx because it is a slice of values, and iterating
		// over such will re-use the same address, making all pointers pointing into it point to the same address
		for i := range block.TransactionVerboseData {
			transaction := block.TransactionVerboseData[i]
			transactionHashesToTxsWithMetadata[transaction.Hash] = &txWithMetadata{
				verboseTx: transaction,
			}
		}
	}

	transactionHashes := transactionHashesToTxsWithMetadataToTransactionHashes(transactionHashesToTxsWithMetadata)

	dbTransactions, err := dbaccess.TransactionsByHashes(dbTx, transactionHashes)
	if err != nil {
		return nil, err
	}

	for _, dbTransaction := range dbTransactions {
		transactionHashesToTxsWithMetadata[dbTransaction.TransactionHash].id = dbTransaction.ID
	}

	newTransactionHashes := make([]string, 0)
	for hash, transaction := range transactionHashesToTxsWithMetadata {
		if transaction.id != 0 {
			continue
		}
		newTransactionHashes = append(newTransactionHashes, hash)
	}

	transactionsToAdd := make([]interface{}, len(newTransactionHashes))
	for i, hash := range newTransactionHashes {
		verboseTx := transactionHashesToTxsWithMetadata[hash].verboseTx

		payload, err := hex.DecodeString(verboseTx.Payload)
		if err != nil {
			return nil, err
		}

		subnetworkID, ok := subnetworkIDsToIDs[verboseTx.SubnetworkID]
		if !ok {
			return nil, errors.Errorf("couldn't find ID for subnetwork %s", verboseTx.SubnetworkID)
		}

		transactionsToAdd[i] = &dbmodels.Transaction{
			TransactionHash: verboseTx.Hash,
			TransactionID:   verboseTx.TxID,
			LockTime:        serializer.Uint64ToBytes(verboseTx.LockTime),
			SubnetworkID:    subnetworkID,
			Gas:             verboseTx.Gas,
			PayloadHash:     verboseTx.PayloadHash,
			Payload:         payload,
			Version:         verboseTx.Version,
		}
	}

	err = dbaccess.BulkInsert(dbTx, transactionsToAdd)
	if err != nil {
		return nil, err
	}

	dbNewTransactions, err := dbaccess.TransactionsByHashes(dbTx, newTransactionHashes)
	if err != nil {
		return nil, err
	}

	if len(dbNewTransactions) != len(newTransactionHashes) {
		return nil, errors.New("couldn't add all new transactions")
	}

	for _, dbTransaction := range dbNewTransactions {
		transactionHashesToTxsWithMetadata[dbTransaction.TransactionHash].id = dbTransaction.ID
		transactionHashesToTxsWithMetadata[dbTransaction.TransactionHash].isNew = true
	}

	return transactionHashesToTxsWithMetadata, nil
}

func convertTxRawResultToMsgTx(tx *appmessage.TransactionVerboseData) (*appmessage.MsgTx, error) {
	txIns := make([]*appmessage.TxIn, len(tx.TransactionVerboseInputs))
	for i, txIn := range tx.TransactionVerboseInputs {
		prevTxID, err := externalapi.NewDomainTransactionIDFromString(txIn.TxID)
		if err != nil {
			return nil, err
		}
		signatureScript, err := hex.DecodeString(txIn.ScriptSig.Hex)
		if err != nil {
			return nil, err
		}
		txIns[i] = &appmessage.TxIn{
			PreviousOutpoint: appmessage.Outpoint{
				TxID:  *prevTxID,
				Index: txIn.OutputIndex,
			},
			SignatureScript: signatureScript,
			Sequence:        txIn.Sequence,
		}
	}
	txOuts := make([]*appmessage.TxOut, len(tx.TransactionVerboseOutputs))
	for i, txOut := range tx.TransactionVerboseOutputs {
		scriptPubKey, err := hex.DecodeString(txOut.ScriptPubKey.Hex)
		if err != nil {
			return nil, err
		}
		txOuts[i] = &appmessage.TxOut{
			Value:        txOut.Value,
			ScriptPubKey: &externalapi.ScriptPublicKey{
				Script:  scriptPubKey,
				Version: 0, // TODO: Update it with real version
			},
		}
	}
	subnetworkID, err := subnetworks.FromString(tx.SubnetworkID)
	if err != nil {
		return nil, err
	}
	if subnetworkID.Equal(&subnetworks.SubnetworkIDNative) {
		return appmessage.NewNativeMsgTx(tx.Version, txIns, txOuts), nil
	}
	payload, err := hex.DecodeString(tx.Payload)
	if err != nil {
		return nil, err
	}
	return appmessage.NewSubnetworkMsgTx(tx.Version, txIns, txOuts, subnetworkID, tx.Gas, payload), nil
}
