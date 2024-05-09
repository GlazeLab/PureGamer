package superadmin

import (
	"github.com/GlazeLab/PureGamer/src/model"
	logging "github.com/ipfs/go-log/v2"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

var log = logging.Logger("superadmin")

const topicName = "/PureGamer/superadmin"

type SuperAdmin struct {
	n   *model.Node
	sub *pubsub.Subscription
	top *pubsub.Topic
}

func NewSuperAdmin(node *model.Node) (*SuperAdmin, error) {
	validator, err := newValidator(node.FixedConfig)
	if err != nil {
		return nil, err
	}
	err = node.PubSub.RegisterTopicValidator(topicName, validator)
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
	return &SuperAdmin{
		n:   node,
		sub: subscription,
		top: topic,
	}, nil
}
