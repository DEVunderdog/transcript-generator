package utils

import "github.com/spf13/viper"

type Config struct {
	Port          string `mapstructure:"PORT"`
	DBSource      string `mapstructure:"DB_SOURCE"`
	Passphrase    string `mapstructure:"PASSPHRASE"`
	Audience      string `mapstructure:"AUDIENCE"`
	Issuer        string `mapstructure:"ISSUER"`
	BucketName    string `mapstructure:"BUCKET_NAME"`
	Domain        string `mapstructure:"DOMAIN"`
	TokenType     string `mapstructure:"TOKEN_TYPE"`
	TokenDuration int    `mapstructure:"TOKEN_DURATION"`
	KeysPurpose   string `mapstructure:"KEYS_PURPOSE"`
	TopicID       string `mapstructure:"TOPIC_ID"`
	ProjectID     string `mapstructure:"PROJECT_ID"`
}

func LoadConfig(path string) (config *Config, err error) {
	viper.SetConfigFile(path)

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return
	}

	return
}
