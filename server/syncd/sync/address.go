package sync

import (
	"github.com/pkg/errors"
	"github.com/someone235/katnip/server/database"
	"github.com/someone235/katnip/server/dbaccess"
	"github.com/someone235/katnip/server/dbmodels"
)

func insertAddresses(dbTx *database.TxContext, transactionHashesToTxsWithMetadata map[string]*txWithMetadata) (map[string]uint64, error) {
	addressSet := make(map[string]struct{})
	for _, transaction := range transactionHashesToTxsWithMetadata {
		if !transaction.isNew {
			continue
		}
		for _, txOut := range transaction.tx.Outputs {
			if txOut.VerboseData.ScriptPublicKeyAddress == "" {
				continue
			}
			addressSet[txOut.VerboseData.ScriptPublicKeyAddress] = struct{}{}
		}
	}
	addresses := stringsSetToSlice(addressSet)

	dbAddresses, err := dbaccess.AddressesByAddressStrings(dbTx, addresses)
	if err != nil {
		return nil, err
	}

	addressesToAddressIDs := make(map[string]uint64)
	for _, dbAddress := range dbAddresses {
		addressesToAddressIDs[dbAddress.Address] = dbAddress.ID
	}

	newAddresses := make([]string, 0)
	for address := range addressSet {
		if _, exists := addressesToAddressIDs[address]; exists {
			continue
		}
		newAddresses = append(newAddresses, address)
	}

	addressesToAdd := make([]interface{}, len(newAddresses))
	for i, address := range newAddresses {
		addressesToAdd[i] = &dbmodels.Address{
			Address: address,
		}
	}

	err = dbaccess.BulkInsert(dbTx, addressesToAdd)
	if err != nil {
		return nil, err
	}

	dbNewAddresses, err := dbaccess.AddressesByAddressStrings(dbTx, newAddresses)
	if err != nil {
		return nil, err
	}
	if len(dbNewAddresses) != len(newAddresses) {
		return nil, errors.New("couldn't add all new addresses")
	}

	for _, dbNewAddress := range dbNewAddresses {
		addressesToAddressIDs[dbNewAddress.Address] = dbNewAddress.ID
	}
	return addressesToAddressIDs, nil
}
