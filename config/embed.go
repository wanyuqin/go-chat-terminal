package config

import (
	"embed"

	"gopkg.in/yaml.v3"
)

//go:embed gchat.yaml
var gchatConfig embed.FS

func LoadConfig() error {
	f, err := gchatConfig.ReadFile("gchat.yaml")
	if err != nil {
		return err
	}
	config := GetConfig()
	return yaml.Unmarshal(f, config)
}
