package exit

import (
	"fmt"
	"github.com/GlazeLab/PureGamer/src/model"
	logging "github.com/ipfs/go-log/v2"
	"io"
	"net"
	"time"
)

var log = logging.Logger("exit")

type Exit struct {
	gameMap map[string]model.Game
}

func NewExit(n *model.Node) (*Exit, error) {
	exitNode := &Exit{
		gameMap: make(map[string]model.Game),
	}
	for _, game := range n.Config.Games {
		exitNode.gameMap[game.ID] = game
	}
	n.FlushConfigCallbacks = append(n.FlushConfigCallbacks, func(c model.Config) error {
		exitNode.gameMap = make(map[string]model.Game)
		for _, game := range c.Games {
			exitNode.gameMap[game.ID] = game
		}
		return nil
	})
	return exitNode, nil
}

func (e *Exit) Handle(s io.ReadWriteCloser, gameId string, extraSend interface{}) error {
	game, ok := e.gameMap[gameId]
	if !ok {
		return nil
	}
	defer s.Close()
	switch game.Protocol {
	case "TCP", "HAProxy":
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", game.Host, game.Port), time.Duration(10)*time.Second)
		if err != nil {
			return err
		}
		defer conn.Close()

		extraFunc, ok := extraSend.(func(conn net.Conn) error)
		if ok {
			err = extraFunc(conn)
			if err != nil {
				return err
			}
		}

		done := make(chan struct{})

		go func() {
			_, err = io.Copy(conn, s)
			if err != nil {
				log.Warn(err)
			}
			done <- struct{}{}
		}()

		go func() {
			_, err = io.Copy(s, conn)
			if err != nil {
				log.Warn(err)
			}
			done <- struct{}{}
		}()

		<-done
		break
	}

	return nil
}
