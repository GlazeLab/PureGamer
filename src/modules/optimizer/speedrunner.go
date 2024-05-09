package optimizer

import (
	"context"
	"github.com/GlazeLab/PureGamer/src/model"
	"github.com/GlazeLab/PureGamer/src/modules/pinging"
	"github.com/GlazeLab/PureGamer/src/utils"
	"github.com/libp2p/go-libp2p/core/peer"
	"math/rand"
	"sync"
)

func speedTestPeers(ctx context.Context, node *model.Node, latencies model.Latencies) {
	connectedPeers := node.Host.Network().Peers()
	var wg sync.WaitGroup
	for _, peerId := range connectedPeers {
		if _, ok := node.BootstrapNodesCheck[peerId]; ok {
			continue
		}
		wg.Add(1)
		go func(peerId peer.ID) {
			defer wg.Done()
			ping, err := pinging.Ping(ctx, node.Host, peerId)
			if err != nil {
				log.Warn(err)
				return
			}
			latencies[peerId.String()] = float64(ping.Milliseconds())
		}(peerId)
	}
	wg.Wait()
	if len(latencies) == 0 {
		log.Warn("No connected peers")
		// need to connect more peers
		peers := node.Host.Peerstore().Peers()
		for i := 0; i < 5; i++ {
			randPicked := peers[rand.Intn(len(peers))]
			if randPicked == node.Host.ID() {
				continue
			}
			node.Host.Network().ConnsToPeer(randPicked)
		}
	}
}

func speedTestGames(ctx context.Context, node *model.Node, latencies model.Latencies) {
	var wg sync.WaitGroup
	for _, game := range node.Config.Games {
		if utils.IsNotAllowed(node.Host.ID().String(), game.ExitNode) {
			continue
		}
		wg.Add(1)
		go func(game model.Game) {
			defer wg.Done()
			switch game.SpeedTestProtocol {
			case "ICMP":
				ping, err := pinging.PingICMP(game.Host, game.Port)
				if err != nil {
					log.Warn(err)
					return
				}
				latencies[game.ID] = float64(ping.Milliseconds())
				break
			case "TCP":
				ping, err := pinging.PingTCP(game.Host, game.Port)
				if err != nil {
					log.Warn(err)
					return
				}
				latencies[game.ID] = float64(ping.Milliseconds())
				break
			}
		}(game)
	}
	wg.Wait()
}

// SpeedTest ping all nodes and measure the latency
func SpeedTest(ctx context.Context, node *model.Node) model.Latencies {
	latencies := make(model.Latencies)
	speedTestPeers(ctx, node, latencies)
	speedTestGames(ctx, node, latencies)
	return latencies
}
