package utils

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Port          string `mapstructure:"SERVER_PORT"`
	DBSource      string `mapstructure:"DB_SOURCE"`
	Passphrase    string `mapstructure:"PASSPHRASE"`
	Audience      string `mapstructure:"AUDIENCE"`
	Issuer        string `mapstructure:"ISSUER"`
	BucketName    string `mapstructure:"BUCKET_NAME"`
	TokenType     string `mapstructure:"TOKEN_TYPE"`
	TokenDuration int    `mapstructure:"TOKEN_DURATION"`
	KeysPurpose   string `mapstructure:"KEYS_PURPOSE"`
	TopicID       string `mapstructure:"TOPIC_ID"`
	ProjectID     string `mapstructure:"PROJECT_ID"`
}

func LoadProdConfig() (config *Config, err error) {
	viper.AutomaticEnv()

	required := []string{
		"PORT",
		"DB_SOURCE",
		"PASSPHRASE",
		"AUDIENCE",
		"ISSUER",
		"BUCKET_NAME",
		"TOKEN_TYPE",
		"TOKEN_DURATION",
		"KEYS_PURPOSE",
		"TOPIC_ID",
		"PROJECT_ID",
	}

	for _, v := range required {
		if !viper.IsSet(v) {
			return nil, fmt.Errorf("missing required environment variable: %s", v)
		}
	}

	config = &Config{}

	err = viper.Unmarshal(&config)

	return
}
