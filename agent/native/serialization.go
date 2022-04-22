package native

import (
	"encoding/json"
	"os"

	"github.com/pelletier/go-toml"
)

func tomlWrite(config interface{}, targetPath string) error {
	file, err := os.Create(targetPath)
	if err != nil {
		return err
	}

	configBytes, err := toml.Marshal(config)

	if err != nil {
		return err
	}

	_, err = file.Write(configBytes)
	return err
}

func jsonWrite(config interface{}, targetPath string) error {
	file, err := os.Create(targetPath)
	if err != nil {
		return err
	}

	configBytes, err := json.Marshal(config)

	if err != nil {
		return err
	}

	_, err = file.Write(configBytes)
	return err
}
