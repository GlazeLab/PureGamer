package pinging

import (
	"github.com/GlazeLab/PureGamer/src/model"
	logging "github.com/ipfs/go-log/v2"
	"time"
)

var log = logging.Logger("ping")

const (
	protocol    = "/PureGamer/ping"
	PingSize    = 32
	pingTimeout = time.Second * 60
	ServiceName = "PureGamer.ping"
)

func Register(node *model.Node) error {
	node.Host.SetStreamHandler(protocol, pingHandler)
	return nil
}
