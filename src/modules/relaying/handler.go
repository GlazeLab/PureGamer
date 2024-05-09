package relaying

import (
	"fmt"
	"github.com/GlazeLab/PureGamer/src/model"
	"github.com/GlazeLab/PureGamer/src/modules/exit"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"io"
	"strings"
)

func getHandler(node *model.Node, exits *exit.Exit) func(network.Stream) {
	return func(income network.Stream) {
		defer income.Close()
		income.Scope().SetService(ServiceName)

		log.Info("Relaying income stream from peer:", income.Conn().RemotePeer().String())

		p := income.Protocol()
		subMatches := pattern.FindStringSubmatch(string(p))
		gameID := subMatches[2]
		nextPeers := strings.Split(subMatches[3], "/next/")

		if len(nextPeers) > 1 {
			log.Info("Relay to peer:", nextPeers[1])
			// forward to the next peer
			nextPeerId, err := peer.Decode(nextPeers[1])
			if err != nil {
				log.Error(err)
				income.Reset()
				return
			}

			// rebuild the protocol id
			protocolId := fmt.Sprintf("/PureGamer/relay/%s/game/%s", subMatches[1], gameID)
			for _, peerId := range nextPeers[2:] {
				protocolId += fmt.Sprintf("/next/%s", peerId)
			}

			outcome, err := node.Host.NewStream(node.CTX, nextPeerId, protocol.ID(protocolId))
			if err != nil {
				log.Error(err)
				income.Reset()
				return
			}
			defer outcome.Close()
			outcome.Scope().SetService(ServiceName)

			done := make(chan struct{})

			go func() {
				_, err = io.Copy(outcome, income)
				if err != nil {
					log.Error(err)
				}
				done <- struct{}{}
			}()

			go func() {
				_, err = io.Copy(income, outcome)
				if err != nil {
					log.Error(err)
				}
				done <- struct{}{}
			}()

			<-done
		} else {
			// final destination
			log.Info("Relay reach final destination")
			err := exits.Handle(income, gameID, nil)
			if err != nil {
				log.Error(err)
				income.Reset()
				return
			}
		}
	}
}
