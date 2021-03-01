package mqtt

import (
	"github.com/someone235/katnip/server/database"
	"path"

	"github.com/someone235/katnip/server/apimodels"
	"github.com/someone235/katnip/server/dbaccess"
	"github.com/someone235/katnip/server/dbmodels"
)

const (
	// TransactionsTopic is an MQTT topic for transactions
	TransactionsTopic = "transactions"

	// AcceptedTransactionsTopic is an MQTT topic for accepted transactions
	AcceptedTransactionsTopic = "transactions/accepted"

	// UnacceptedTransactionsTopic is an MQTT topic for unaccepted transactions
	UnacceptedTransactionsTopic = "transactions/unaccepted"
)

// publishTransactionsNotifications publishes notifications for each transaction of the given transactions
func publishTransactionsNotifications(topic string, dbTransactions []*dbmodels.Transaction, selectedTipBlueScore uint64) error {
	for _, dbTransaction := range dbTransactions {
		transaction := apimodels.ConvertTxModelToTxResponse(dbTransaction, selectedTipBlueScore)
		addresses := uniqueAddressesForTransaction(transaction)
		for _, address := range addresses {
			err := publishTransactionNotificationForAddress(transaction, address, topic)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func uniqueAddressesForTransaction(transaction *apimodels.TransactionResponse) []string {
	addressesMap := make(map[string]struct{})
	addresses := []string{}
	for _, output := range transaction.Outputs {
		if _, exists := addressesMap[output.Address]; !exists {
			addresses = append(addresses, output.Address)
			addressesMap[output.Address] = struct{}{}
		}
	}
	for _, input := range transaction.Inputs {
		if _, exists := addressesMap[input.Address]; !exists {
			addresses = append(addresses, input.Address)
			addressesMap[input.Address] = struct{}{}
		}
	}
	return addresses
}

func publishTransactionNotificationForAddress(transaction *apimodels.TransactionResponse, address string, topic string) error {
	return publish(path.Join(topic, address), transaction)
}

// PublishUnacceptedTransactionsNotifications publishes notification for each unaccepted transaction of the given chain-block
func PublishUnacceptedTransactionsNotifications(unacceptedTransactions []*dbmodels.Transaction) error {
	if !isConnected() {
		return nil
	}

	selectedTipBlueScore, err := dbaccess.SelectedTipBlueScore(database.NoTx())
	if err != nil {
		return err
	}

	err = publishTransactionsNotifications(UnacceptedTransactionsTopic, unacceptedTransactions, selectedTipBlueScore)
	if err != nil {
		return err
	}
	return nil
}
