package utils

import (
	"encoding/json"
	"github.com/GlazeLab/PureGamer/src/model"
	"io"
	"os"
)

func LoadConfig(path string) (*model.Config, *model.FixedConfig, error) {
	fixedConfigFile, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer fixedConfigFile.Close()
	decoder := json.NewDecoder(fixedConfigFile)
	var fixedConfig model.FixedConfig
	err = decoder.Decode(&fixedConfig)
	if err != nil {
		return nil, nil, err
	}

	configFile, err := os.Open(fixedConfig.ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			configFile, err = os.Create(fixedConfig.ConfigPath)
			if err != nil {
				return nil, nil, err
			}
			config := model.Config{
				Games:  []model.Game{},
				System: model.System{ListenHost: ""},
			}
			encoder := json.NewEncoder(configFile)
			err = encoder.Encode(config)
			if err != nil {
				return nil, nil, err
			}
			return &config, &fixedConfig, nil
		} else {
			return nil, nil, err
		}
	} else {
		defer configFile.Close()
		decoder = json.NewDecoder(configFile)
		var config model.Config
		err = decoder.Decode(&config)
		if err != nil {
			return nil, nil, err
		}
		return &config, &fixedConfig, nil
	}
}

func ReadText(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	bytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
