package relaying

import (
	"context"
	"fmt"
	"github.com/GlazeLab/PureGamer/src/model"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

func OpenRelay(ctx context.Context, n *model.Node, gameId string, relayNodes []string) (network.Stream, error) {
	// the first node is what you want to connect to
	// the last node should be the exit node
	var firstRelayPeerId peer.ID
	var err error
	if len(relayNodes) == 0 {
		firstRelayPeerId = n.Host.ID()
	} else {
		firstRelayPeerId, err = peer.Decode(relayNodes[0])
		if err != nil {
			return nil, err
		}
	}

	// build the protocol id
	protocolId := fmt.Sprintf("/PureGamer/relay/%s/game/%s", version, gameId)
	if len(relayNodes) > 1 {
		for _, peerId := range relayNodes[1:] {
			protocolId += fmt.Sprintf("/next/%s", peerId)
		}
	}

	relayStream, err := n.Host.NewStream(context.TODO(), firstRelayPeerId, protocol.ID(protocolId))
	if err != nil {
		return nil, err
	}
	relayStream.Scope().SetService(ServiceName)

	return relayStream, nil
}
