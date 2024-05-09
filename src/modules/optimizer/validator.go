package optimizer

import (
	"context"
	"github.com/GlazeLab/PureGamer/src/model"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/vmihailenco/msgpack/v5"
)

func validator(ctx context.Context, pid peer.ID, msg *pubsub.Message) bool {
	var cmd model.Latencies
	err := msgpack.Unmarshal(msg.GetData(), &cmd)
	if err != nil {
		return false
	}
	return true
}
