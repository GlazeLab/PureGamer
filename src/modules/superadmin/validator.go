package superadmin

import (
	"context"
	"github.com/GlazeLab/PureGamer/src/model"
	"github.com/GlazeLab/PureGamer/src/utils"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/vmihailenco/msgpack/v5"
)

func newValidator(config *model.FixedConfig) (interface{}, error) {
	pubKey, err := utils.DecodePublic(config.SuperAdminPubKey)
	if err != nil {
		return nil, err
	}
	return func(ctx context.Context, p peer.ID, msg *pubsub.Message) bool {
		var signedConfig model.SignedConfig
		err := msgpack.Unmarshal(msg.GetData(), &signedConfig)
		if err != nil {
			log.Error(err)
			return false
		}
		log.Info("Received config from superadmin")
		sign := signedConfig.Sign
		configBytes, err := msgpack.Marshal(signedConfig.Config)
		if err != nil {
			log.Error(err)
			return false
		}
		return utils.Verify(configBytes, sign, pubKey)
	}, nil
}
