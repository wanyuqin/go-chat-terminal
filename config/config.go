package config

import "sync"

type Config struct {
	logo string

	OpenAIKey string

	Port int64
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
