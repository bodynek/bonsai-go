package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	NoAPI     bool   `mapstructure:"no_api"`
	APIPort   int    `mapstructure:"api_port"`
	SvcPort   int    `mapstructure:"svc_port"`
	RedisHost string `mapstructure:"redis_host"`
	RedisPort int    `mapstructure:"redis_port"`
}

func LoadConfig(configFile string) (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("no_api", false)
	v.SetDefault("api_port", 8081)
	v.SetDefault("svc_port", 8080)
	v.SetDefault("redis_host", "localhost")
	v.SetDefault("redis_port", 6379)

	// Read from environment variables
	v.AutomaticEnv()

	// Read from configuration file, if specified
	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal configuration into the struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &cfg, nil
}
