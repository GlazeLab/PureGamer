package entry

import (
	"fmt"
	"github.com/GlazeLab/PureGamer/src/model"
	"github.com/GlazeLab/PureGamer/src/modules/exit"
	"github.com/GlazeLab/PureGamer/src/modules/optimizer"
	"github.com/GlazeLab/PureGamer/src/modules/relaying"
	"github.com/GlazeLab/PureGamer/src/utils"
	logging "github.com/ipfs/go-log/v2"
	"github.com/pires/go-proxyproto"
	"io"
	"net"
)

var log = logging.Logger("entry")

func Listen(n *model.Node, exits *exit.Exit, optimize *optimizer.Optimizer) error {
	gameMap := make(map[string]model.Game)
	listening := make(map[string]interface{})

	for _, game := range n.Config.Games {
		gameMap[game.ID] = game
	}
	nodeId := n.Host.ID().String()

	listen := func(c model.Config) error {
		for _, game := range c.Games {
			if utils.IsNotAllowed(nodeId, game.EntryNode) {
				continue
			}
			switch game.Protocol {
			case "TCP", "HAProxy":
				listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", c.System.ListenHost, game.ListenPort))
				log.Info(fmt.Sprintf("Listening on %s:%d", c.System.ListenHost, game.ListenPort))
				if err != nil {
					return err
				}
				listening[game.ID] = listener
				break
			case "UDP":
				listener, err := net.ListenPacket("udp", fmt.Sprintf("%s:%d", c.System.ListenHost, game.ListenPort))
				log.Info(fmt.Sprintf("Listening on %s:%d", c.System.ListenHost, game.ListenPort))
				if err != nil {
					return err
				}
				listening[game.ID] = listener
				break
			}
		}
		for gameId, listener := range listening {
			listener := listener
			gameId := gameId
			switch gameMap[gameId].Protocol {
			case "TCP":
				go func() {
					for {
						income, err := listener.(net.Listener).Accept()
						if err != nil {
							log.Error(err)
							return
						}
						func() {
							defer income.Close()
							relayNodes := optimize.OptimizedRoutes(nodeId, gameId)
							log.Infof("Relays: %v", relayNodes)

							if len(relayNodes) == 0 || relayNodes[len(relayNodes)-1] == nodeId {
								if utils.IsNotAllowed(nodeId, gameMap[gameId].ExitNode) {
									log.Warnf("This is not allowed to relay to %s", gameId)
									return
								}
								err := exits.Handle(income, gameId, nil)
								if err != nil {
									log.Error(err)
									return
								}
							} else {
								relay, err := relaying.OpenRelay(n.CTX, n, gameId, relayNodes)
								if err != nil {
									log.Error(err)
									return
								}
								defer relay.Close()

								done := make(chan struct{})

								go func() {
									_, err = io.Copy(income, relay)
									if err != nil {
										log.Warn(err)
									}
									done <- struct{}{}
								}()

								go func() {
									_, err = io.Copy(relay, income)
									if err != nil {
										log.Warn(err)
									}
									done <- struct{}{}
								}()
								<-done
							}
						}()
					}
				}()
				break
			case "HAProxy":
				go func() {
					for {
						income, err := listener.(net.Listener).Accept()

						if err != nil {
							log.Error(err)
							return
						}
						func() {
							defer income.Close()
							log.Info(income.RemoteAddr().(*net.TCPAddr).IP)
							relayNodes := optimize.OptimizedRoutes(nodeId, gameId)
							proxyHeader := proxyproto.HeaderProxyFromAddrs(2, income.RemoteAddr(), income.LocalAddr())
							log.Infof("Relays: %v", relayNodes)

							if len(relayNodes) == 0 || relayNodes[len(relayNodes)-1] == nodeId {
								if utils.IsNotAllowed(nodeId, gameMap[gameId].ExitNode) {
									log.Warnf("This is not allowed to relay to %s", gameId)
									return
								}
								err = exits.Handle(income, gameId, func(conn net.Conn) error {
									_, err := proxyHeader.WriteTo(conn)
									return err
								})
								if err != nil {
									log.Error(err)
									return
								}
							} else {
								relay, err := relaying.OpenRelay(n.CTX, n, gameId, relayNodes)
								if err != nil {
									log.Error(err)
									return
								}
								defer relay.Close()

								_, err = proxyHeader.WriteTo(relay)
								if err != nil {
									log.Error(err)
									return
								}

								done := make(chan struct{})

								go func() {
									_, err = io.Copy(income, relay)
									if err != nil {
										log.Warn(err)
									}
									done <- struct{}{}
								}()

								go func() {
									_, err = io.Copy(relay, income)
									if err != nil {
										log.Warn(err)
									}
									done <- struct{}{}
								}()
								<-done
							}
						}()
					}
				}()
				break
			case "UDP":
				/*
					established := make(map[string]network.Stream)
					go func() {
						for {
							buf := make([]byte, 1024)
							n, addr, err := listener.(net.PacketConn).ReadFrom(buf)
							if err != nil {
								log.Error(err)
								return
							}

						}

					}()
					// To be implemented
				*/
				break
			}
		}
		return nil
	}

	n.FlushConfigCallbacks = append(n.FlushConfigCallbacks, func(c model.Config) error {
		gameMap = make(map[string]model.Game)
		for _, game := range c.Games {
			gameMap[game.ID] = game
		}
		for _, listener := range listening {
			switch listener.(type) {
			case net.Listener:
				listener.(net.Listener).Close()
				break
			case net.PacketConn:
				listener.(net.PacketConn).Close()
				break
			}
		}
		listening = make(map[string]interface{})
		err := listen(c)
		if err != nil {
			return err
		}
		return nil
	})

	err := listen(*n.Config)
	if err != nil {
		return err
	}

	return nil
}
