package node

import (
	"context"
	"errors"
	"fmt"
	"github.com/GlazeLab/PureGamer/src/model"
	"github.com/GlazeLab/PureGamer/src/modules/pinging"
	"github.com/GlazeLab/PureGamer/src/utils"
	dbstore "github.com/ipfs/go-ds-leveldb"
	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/metrics"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/core/routing"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/libp2p/go-libp2p/p2p/host/peerstore/pstoreds"
	"github.com/libp2p/go-libp2p/p2p/muxer/yamux"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	"io"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

var log = logging.Logger("node")

const sn = "PureGamerNetwork-0"

func getBootstrapPeers(peers []string) []peer.AddrInfo {
	addrs := make([]peer.AddrInfo, len(peers))
	var temp *peer.AddrInfo
	var err error
	for i, pstring := range peers {
		temp, err = peer.AddrInfoFromString(pstring)
		if err != nil {
			panic(err)
		}
		addrs[i] = *temp
	}
	return addrs
}

func Listen(ctx context.Context) (*model.Node, error) {
	config, fixedConfig, err := utils.LoadConfig("config.json")
	if err != nil {
		return nil, err
	}

	var priv crypto.PrivKey
	var privateKeyFile *os.File

	privateKeyFile, err = os.Open(path.Join(fixedConfig.DataPath, "priv.key"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			priv, _, err = crypto.GenerateKeyPair(
				crypto.Ed25519,
				-1,
			)
			// save the private key to a file
			var privBytes []byte
			privBytes, err = crypto.MarshalPrivateKey(priv)
			if err != nil {
				return nil, err
			}
			log.Info("Generated private key")
			if _, err = os.Stat(fixedConfig.DataPath); os.IsNotExist(err) {
				err = os.MkdirAll(fixedConfig.DataPath, 0666)
				if err != nil {
					return nil, err
				}
			}
			privateKeyFile, err = os.Create(path.Join(fixedConfig.DataPath, "priv.key"))
			if err != nil {
				return nil, err
			}
			_, err = privateKeyFile.Write(privBytes)
			err = privateKeyFile.Close()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		var privBytes []byte
		privBytes, err = io.ReadAll(privateKeyFile)
		if err != nil {
			return nil, err
		}
		priv, err = crypto.UnmarshalPrivateKey(privBytes)
		if err != nil {
			return nil, err
		}
		err = privateKeyFile.Close()
		if err != nil {
			return nil, err
		}
	}

	connMgr, err := connmgr.NewConnManager(
		100,
		400,
	)

	store, err := dbstore.NewDatastore(path.Join(fixedConfig.DataPath, "p2p"), nil)

	if err != nil {

		return nil, err
	}

	peerStore, err := pstoreds.NewPeerstore(ctx, store, pstoreds.Options{
		CacheSize:    100,
		MaxProtocols: 5,
	})
	if err != nil {
		return nil, err
	}

	bwReport := metrics.NewBandwidthCounter()

	var idht *dht.IpfsDHT

	bootNodesAddrs := getBootstrapPeers(fixedConfig.BoostrapNodes)

	h, err := libp2p.New(
		libp2p.Peerstore(peerStore),
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", fixedConfig.Port),
			fmt.Sprintf("/ip6/::/udp/%d/quic-v1", fixedConfig.Port),
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", fixedConfig.Port),
			fmt.Sprintf("/ip6/::/tcp/%d", fixedConfig.Port),
		),
		libp2p.Security(noise.ID, noise.New),
		libp2p.Identity(priv),
		libp2p.DefaultTransports,
		libp2p.Muxer(yamux.ID, yamux.DefaultTransport),
		libp2p.ConnectionManager(connMgr),
		libp2p.DefaultResourceManager,
		libp2p.DefaultMultiaddrResolver,
		libp2p.Ping(false),
		libp2p.DisableRelay(),
		libp2p.BandwidthReporter(bwReport),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			idht, err = dht.New(ctx, h,
				dht.Mode(dht.ModeServer),
				dht.BootstrapPeers(bootNodesAddrs...),
				dht.Datastore(store),
				dht.ProtocolPrefix("/PureGamer"),
				dht.BucketSize(8),
			)
			if err != nil {
				panic(err)
			}
			return idht, err
		}),
		libp2p.WithDialTimeout(time.Second*10),
	)
	if err != nil {
		return nil, err
	}

	// print the node's listening addresses
	fmt.Println("Listen addresses:", h.Addrs())

	// print the node's ID
	fmt.Println("Node ID:", h.ID())

	// print the node's public key
	fmt.Println("Public Key:", h.Peerstore().PubKey(h.ID()))

	// connect to the bootstrap nodes
	bootNodeCheckMap := make(map[peer.ID]struct{})
	var wg sync.WaitGroup
	for _, addr := range bootNodesAddrs {
		wg.Add(1)
		go func(addr peer.AddrInfo) {
			defer wg.Done()
			if err = h.Connect(ctx, addr); err != nil {
				log.Error(err)
			}
			if _, err = h.Network().DialPeer(ctx, addr.ID); err != nil {
				log.Error(err)
			}
			h.Peerstore().AddAddrs(addr.ID, addr.Addrs, peerstore.PermanentAddrTTL)
			if _, err = idht.RoutingTable().TryAddPeer(addr.ID, true, false); err != nil {
				log.Error(err)
			}

			_, err := pinging.Ping(ctx, h, addr.ID)
			if err != nil {
				log.Warn(err)
				bootNodeCheckMap[addr.ID] = struct{}{}
				return
			}
		}(addr)
	}
	wg.Wait()

	node := &model.Node{
		Host:                 h,
		Router:               idht,
		CTX:                  ctx,
		Store:                store,
		Config:               config,
		FixedConfig:          fixedConfig,
		FlushConfigCallbacks: make([]func(model.Config) error, 0),
		BootstrapNodesCheck:  bootNodeCheckMap,
	}

	if err != nil {
		return nil, err
	}

	routingDiscovery := drouting.NewRoutingDiscovery(node.Router)

	go func(serviceName string) {
		ticker := time.NewTicker(time.Second * 60)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				dutil.Advertise(node.CTX, routingDiscovery, serviceName)
				break
			case <-ctx.Done():
				return
			}
		}
	}(sn)

	go func(serviceName string) {
		ticker := time.NewTicker(time.Second * 10)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				peerChan, err := routingDiscovery.FindPeers(node.CTX, serviceName)
				if err != nil {
					log.Error(err)
					return
				}

				for peer := range peerChan {
					if peer.ID == node.Host.ID() {
						continue
					}
					// check if already connected
					if node.Host.Peerstore().PeerInfo(peer.ID).Addrs != nil {
						continue
					}
					if err := node.Host.Connect(node.CTX, peer); err != nil {
						log.Error(err)
					}

					remotePeerId := peer.ID

					latency, err := pinging.Ping(node.CTX, node.Host, remotePeerId)
					if err != nil {
						if strings.Contains(err.Error(), "failed to negotiate protocol: protocols not supported") {
							log.Warn("Not a PureGamer Node: ", remotePeerId)
							node.Host.Network().ClosePeer(remotePeerId)
							node.Host.Peerstore().RemovePeer(remotePeerId)
							node.Host.Peerstore().ClearAddrs(remotePeerId)
							node.PubSub.BlacklistPeer(remotePeerId)
							node.Router.RoutingTable().RemovePeer(remotePeerId)
						} else {
							log.Error(err)
						}
					} else {
						log.Info("Discovered: ", peer, " with latency: ", latency)
					}
				}
				break
			case <-ctx.Done():
				return
			}
		}
	}(sn)
	err = idht.Bootstrap(ctx)
	if err != nil {
		return nil, err
	}

	ps, err := pubsub.NewGossipSub(ctx, h,
		pubsub.WithDiscovery(drouting.NewRoutingDiscovery(idht)),
		pubsub.WithFloodPublish(true),
		pubsub.WithMessageSignaturePolicy(pubsub.StrictSign),
		pubsub.WithPeerExchange(true),
	)

	if err != nil {
		return nil, err
	}
	node.PubSub = ps

	return node, nil
}
