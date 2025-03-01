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
	viper.SetEnvPrefix("")
	viper.AutomaticEnv()

	config = &Config{}

	viper.BindEnv("SERVER_PORT")
	viper.BindEnv("DB_SOURCE")
	viper.BindEnv("PASSPHRASE")
	viper.BindEnv("AUDIENCE")
	viper.BindEnv("ISSUER")
	viper.BindEnv("BUCKET_NAME")
	viper.BindEnv("TOKEN_TYPE")
	viper.BindEnv("TOKEN_DURATION")
	viper.BindEnv("KEYS_PURPOSE")
	viper.BindEnv("TOPIC_ID")
	viper.BindEnv("PROJECT_ID")

	required := []string{
		"SERVER_PORT",
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

	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %v", err)
	}
	
	return
}
