package relaying

import (
	"github.com/GlazeLab/PureGamer/src/model"
	"github.com/GlazeLab/PureGamer/src/modules/exit"
	logging "github.com/ipfs/go-log/v2"
	"regexp"
)

var (
	pattern, _ = regexp.Compile(`^/PureGamer/relay/([a-zA-Z0-9.]+)/game/([123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]+)((?:/next/[123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]+)*)$`)
	log        = logging.Logger("relaying")
)

const (
	ServiceName = "PureGamer.relay"
	version     = "0.0.1"
)

// Register /PureGamer/relay/<version>/game/<gameid>/next/<peerid>/next/<peerid>/...
func Register(node *model.Node, exits *exit.Exit) error {
	node.Host.SetStreamHandlerMatch("/PureGamer/relay/"+version, match, getHandler(node, exits))
	return nil
}
