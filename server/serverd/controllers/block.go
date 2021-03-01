package controllers

import (
	"encoding/hex"
	"github.com/kaspanet/kaspad/domain/consensus/model/externalapi"
	"github.com/someone235/katnip/server/database"
	"net/http"

	"github.com/someone235/katnip/server/apimodels"
	"github.com/someone235/katnip/server/dbaccess"
	"github.com/someone235/katnip/server/dbmodels"

	"github.com/pkg/errors"

	"github.com/someone235/katnip/server/httpserverutils"
)

const maxGetBlocksLimit = 100

// GetBlockByHashHandler returns a block by a given hash.
func GetBlockByHashHandler(blockHash string) (interface{}, error) {
	if bytes, err := hex.DecodeString(blockHash); err != nil || len(bytes) != externalapi.DomainHashSize {
		return nil, httpserverutils.NewHandlerError(http.StatusUnprocessableEntity,
			errors.Errorf("the given block hash is not a hex-encoded %d-byte hash", externalapi.DomainHashSize))
	}

	preloadedFields := append([]dbmodels.FieldName{dbmodels.BlockFieldNames.Transactions},
		dbmodels.BlockRecommendedPreloadedFields...)
	block, err := dbaccess.BlockByHash(database.NoTx(), blockHash, preloadedFields...)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, httpserverutils.NewHandlerError(http.StatusNotFound, errors.New("no block with the given block hash was found"))
	}

	selectedTipBlueScore, err := dbaccess.SelectedTipBlueScore(database.NoTx())
	if err != nil {
		return nil, err
	}

	blockRes := apimodels.ConvertBlockModelToBlockResponse(block, selectedTipBlueScore)
	return blockRes, nil
}

// GetBlocksHandler searches for all blocks
func GetBlocksHandler(orderString string, skip, limit int64) (interface{}, error) {
	if limit > maxGetBlocksLimit || limit < 1 {
		return nil, httpserverutils.NewHandlerError(http.StatusBadRequest,
			errors.Errorf("limit higher than %d or lower than 1 was requested", maxGetBlocksLimit))
	}

	if skip < 0 {
		return nil, httpserverutils.NewHandlerError(http.StatusBadRequest,
			errors.New("skip lower than 0 was requested"))
	}

	order, err := dbaccess.StringToOrder(orderString)
	if err != nil {
		return nil, httpserverutils.NewHandlerError(http.StatusUnprocessableEntity, err)
	}

	blocks, err := dbaccess.Blocks(database.NoTx(), order, uint64(skip), uint64(limit), dbmodels.BlockRecommendedPreloadedFields...)
	if err != nil {
		return nil, err
	}

	selectedTipBlueScore, err := dbaccess.SelectedTipBlueScore(database.NoTx())
	if err != nil {
		return nil, err
	}

	blockResponses := make([]*apimodels.BlockResponse, len(blocks))
	for i, block := range blocks {
		blockResponses[i] = apimodels.ConvertBlockModelToBlockResponse(block, selectedTipBlueScore)
	}

	return blockResponses, nil
}

// GetBlockCountHandler returns the total number of blocks.
func GetBlockCountHandler() (interface{}, error) {
	return dbaccess.BlocksCount(database.NoTx())
}
