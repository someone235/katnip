package sync

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/logger"
	"github.com/someone235/katnip/server/database"
	"github.com/someone235/katnip/server/dbaccess"
	"github.com/someone235/katnip/server/dbmodels"
	"github.com/someone235/katnip/server/kaspadrpc"

	"github.com/pkg/errors"
)

func insertSubnetworks(client *kaspadrpc.Client, dbTx *database.TxContext, blocks []*appmessage.RPCBlock) (
	subnetworkIDsToIDs map[string]uint64, err error) {

	onEnd := logger.LogAndMeasureExecutionTime(log, "insertSubnetworks")
	defer onEnd()

	subnetworkSet := make(map[string]struct{})
	for _, block := range blocks {
		for _, transaction := range block.Transactions {
			subnetworkSet[transaction.SubnetworkID] = struct{}{}
		}
	}

	subnetworkIDs := stringsSetToSlice(subnetworkSet)

	dbSubnetworks, err := dbaccess.SubnetworksByIDs(dbTx, subnetworkIDs)
	if err != nil {
		return nil, err
	}

	subnetworkIDsToIDs = make(map[string]uint64)
	for _, dbSubnetwork := range dbSubnetworks {
		subnetworkIDsToIDs[dbSubnetwork.SubnetworkID] = dbSubnetwork.ID
	}

	newSubnetworkIDs := make([]string, 0)
	for subnetworkID := range subnetworkSet {
		if _, exists := subnetworkIDsToIDs[subnetworkID]; exists {
			continue
		}
		newSubnetworkIDs = append(newSubnetworkIDs, subnetworkID)
	}

	subnetworksToAdd := make([]interface{}, len(newSubnetworkIDs))
	for i, subnetworkID := range newSubnetworkIDs {
		var gasLimit *uint64 // TODO: Fill with real value
		subnetworksToAdd[i] = &dbmodels.Subnetwork{
			SubnetworkID: subnetworkID,
			GasLimit:     gasLimit,
		}
	}

	err = dbaccess.BulkInsert(dbTx, subnetworksToAdd)
	if err != nil {
		return nil, err
	}

	dbNewSubnetworks, err := dbaccess.SubnetworksByIDs(dbTx, newSubnetworkIDs)
	if err != nil {
		return nil, err
	}

	if len(dbNewSubnetworks) != len(newSubnetworkIDs) {
		return nil, errors.New("couldn't add all new subnetworks")
	}

	for _, dbSubnetwork := range dbNewSubnetworks {
		subnetworkIDsToIDs[dbSubnetwork.SubnetworkID] = dbSubnetwork.ID
	}
	return subnetworkIDsToIDs, nil
}
