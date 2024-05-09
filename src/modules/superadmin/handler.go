package superadmin

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"github.com/GlazeLab/PureGamer/src/model"
	"github.com/GlazeLab/PureGamer/src/utils"
	"github.com/vmihailenco/msgpack/v5"
	"os"
)

func (su *SuperAdmin) Handle(ctx context.Context) <-chan error {
	errCh := make(chan error)
	go func() {
		for {
			msg, err := su.sub.Next(ctx)
			if err != nil {
				log.Error(err)
				errCh <- err
				continue
			}
			var signedConfig model.SignedConfig
			err = msgpack.Unmarshal(msg.GetData(), &signedConfig)
			if err != nil {
				log.Error(err)
				errCh <- err
				continue
			}
			err = su.handleConfig(signedConfig)
			if err != nil {
				log.Error(err)
				errCh <- err
				continue
			}
		}
	}()
	return errCh
}

func (su *SuperAdmin) handleConfig(signedConfig model.SignedConfig) error {
	su.n.Config = &signedConfig.Config
	var err error
	for _, cb := range su.n.FlushConfigCallbacks {
		err = cb(signedConfig.Config)
		if err != nil {
			return err
		}
	}
	// write to file
	var file *os.File
	file, err = os.Create(su.n.FixedConfig.ConfigPath)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	err = enc.Encode(signedConfig.Config)
	if err != nil {
		return err
	}
	return nil
}

func (su *SuperAdmin) SendConfig(ctx context.Context, config model.Config, privKey *ecdsa.PrivateKey) error {
	configBytes, err := msgpack.Marshal(config)
	if err != nil {
		return err
	}
	sign, err := utils.Sign(configBytes, privKey)
	if err != nil {
		return err
	}
	signedConfig := model.SignedConfig{
		Config: config,
		Sign:   sign,
	}
	msg, err := msgpack.Marshal(signedConfig)
	if err != nil {
		return err
	}
	err = su.top.Publish(ctx, msg)
	if err != nil {
		return err
	}
	log.Info("Sent config to superadmin")
	return nil
}
