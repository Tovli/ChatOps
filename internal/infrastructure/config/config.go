package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	GitHub   GitHubConfig   `mapstructure:"github"`
	Slack    SlackConfig    `mapstructure:"slack"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type GitHubConfig struct {
	Token string `mapstructure:"token"`
}

type SlackConfig struct {
	BotToken   string `mapstructure:"bot_token"`
	SigningKey string `mapstructure:"signing_key"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set environment variable mappings with CHATOPS_ prefix
	viper.SetEnvPrefix("CHATOPS")
	viper.AutomaticEnv()

	// Map environment variables to config fields
	viper.BindEnv("github.token", "CHATOPS_GITHUB_TOKEN")
	viper.BindEnv("slack.bot_token", "CHATOPS_SLACK_BOT_TOKEN")
	viper.BindEnv("slack.signing_key", "CHATOPS_SLACK_SIGNING_KEY")
	viper.BindEnv("database.host", "CHATOPS_DB_HOST")
	viper.BindEnv("database.port", "CHATOPS_DB_PORT")
	viper.BindEnv("database.user", "CHATOPS_DB_USER")
	viper.BindEnv("database.password", "CHATOPS_DB_PASSWORD")
	viper.BindEnv("database.dbname", "CHATOPS_DB_NAME")
	viper.BindEnv("database.sslmode", "CHATOPS_DB_SSLMODE")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
