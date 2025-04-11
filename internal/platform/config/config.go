package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Slack    SlackConfig
	GitHub   GitHubConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Port int
}

type SlackConfig struct {
	BotToken   string
	AppToken   string
	SigningKey string
}

type GitHubConfig struct {
	AppID          string
	InstallationID string
	PrivateKey     string
}

type DatabaseConfig struct {
	DSN string
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
