package optimizer

import (
	"github.com/GlazeLab/PureGamer/src/model"
	logging "github.com/ipfs/go-log/v2"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

var log = logging.Logger("optimizer")

const topicName = "/PureGamer/latencies"

type Optimizer struct {
	gr  *model.Graph
	sub *pubsub.Subscription
	top *pubsub.Topic
	n   *model.Node
}

func NewOptimizer(node *model.Node) (*Optimizer, error) {
	err := node.PubSub.RegisterTopicValidator(topicName, validator)
	if err != nil {
		return nil, err
	}

	topic, err := node.PubSub.Join(topicName)
	if err != nil {
		return nil, err
	}
	_, err = topic.Relay()
	if err != nil {
		return nil, err
	}
	subscription, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	graph := model.NewGraph()

	optimizer := Optimizer{
		gr:  graph,
		sub: subscription,
		top: topic,
		n:   node,
	}
	return &optimizer, nil
}

func (o *Optimizer) Info() string {
	return o.gr.Print()
}

func (o *Optimizer) RouteInfo(from string, to string) string {
	return o.gr.PrintRoutes(from, to)
}
