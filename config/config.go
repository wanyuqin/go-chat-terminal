package config

import "sync"

type Config struct {
	logo string

	OpenAIKey string `yaml:"open_ai_key"`

	Port int64 `yaml:"port"`
}

var (
	cfg  *Config
	once sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		cfg = &Config{}
	})

	return cfg
}
