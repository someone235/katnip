package kaspadrpc

import (
	"github.com/kaspanet/kaspad/app/appmessage"
	"github.com/kaspanet/kaspad/infrastructure/network/rpcclient"
	"github.com/kaspanet/kasparov/config"
	"time"

	"github.com/pkg/errors"
)

const timeout = 10 * time.Second

// Client represents a connection to the JSON-RPC API of a full node
type Client struct {
	*rpcclient.RPCClient

	OnBlockAdded   chan *appmessage.BlockAddedNotificationMessage
}

var clientInstance *Client

// GetClient returns an instance of the RPC client, in case we have an active connection
func GetClient() (*Client, error) {
	if clientInstance == nil {
		return nil, errors.New("RPC is not connected")
	}

	return clientInstance, nil
}

// NewClient creates a new Client
func NewClient(cfg *config.KasparovFlags, subscribeToNotifications bool) (*Client, error) {
	rpcAddress, err := cfg.NetParams().NormalizeRPCServerAddress(cfg.RPCServer)
	if err != nil {
		return nil, err
	}
	rpcClient, err := rpcclient.NewRPCClient(rpcAddress)
	if err != nil {
		return nil, err
	}
	rpcClient.SetTimeout(timeout)

	const channelCapacity = 1_000_000
	client := &Client{
		RPCClient:      rpcClient,
		OnBlockAdded:   make(chan *appmessage.BlockAddedNotificationMessage, channelCapacity),
	}

	if subscribeToNotifications {
		err = rpcClient.RegisterForBlockAddedNotifications(func(notification *appmessage.BlockAddedNotificationMessage) {
			client.OnBlockAdded <- notification
		})
		if err != nil {
			return nil, errors.Wrapf(err, "error requesting block-added notifications")
		}
	}

	clientInstance = client

	return client, nil
}
