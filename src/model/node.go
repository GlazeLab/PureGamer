package model

import (
	"context"
	dbstore "github.com/ipfs/go-ds-leveldb"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

type Node struct {
	Host                 host.Host
	Router               *dht.IpfsDHT
	CTX                  context.Context
	Store                *dbstore.Datastore
	PubSub               *pubsub.PubSub
	Config               *Config
	FixedConfig          *FixedConfig
	FlushConfigCallbacks []func(Config) error
	BootstrapNodesCheck  map[peer.ID]struct{}
}
