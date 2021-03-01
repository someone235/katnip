package mqtt

import (
	"github.com/someone235/katnip/server/apimodels"
	"github.com/someone235/katnip/server/database"
	"github.com/someone235/katnip/server/dbaccess"
	"github.com/someone235/katnip/server/dbmodels"
)

const (
	// SelectedTipTopic is an MQTT topic for DAG selected tips
	SelectedTipTopic = "dag/selected-tip"
)

// PublishSelectedTipNotification publishes notification for a new selected tip
func PublishSelectedTipNotification(selectedTipHash string) error {
	if !isConnected() {
		return nil
	}
	dbBlock, err := dbaccess.BlockByHash(database.NoTx(), selectedTipHash, dbmodels.BlockRecommendedPreloadedFields...)
	if err != nil {
		return err
	}

	block := apimodels.ConvertBlockModelToBlockResponse(dbBlock, dbBlock.BlueScore)
	return publish(SelectedTipTopic, block)
}
