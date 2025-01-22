package utils

import "github.com/spf13/viper"

type Config struct {
	ClientID              string `mapstructure:"GOOGLE_CLIENT_ID"`
	ClientSecret          string `mapstructure:"GOOGLE_CLIENT_SECRET"`
	Port                  string `mapstructure:"PORT"`
	DBSource              string `mapstructure:"DB_SOURCE"`
	Passphrase            string `mapstructure:"PASSPHRASE"`
	SessionKey            string `mapstructure:"SESSION_KEY"`
	RedirectURL           string `mapstructure:"REDIRECT_URL"`
	Audience              string `mapstructure:"AUDIENCE"`
	Issuer                string `mapstructure:"ISSUER"`
	FrontendURL           string `mapstructure:"FRONTEND_URL"`
	ServiceAccountKeyPath string `mapstructure:"SERVICE_ACCOUNT_KEY_PATH"`
	BucketName            string `mapstructure:"BUCKET_NAME"`
	Domain                string `mapstructure:"DOMAIN"`
	TokenType             string `mapstructure:"TOKEN_TYPE"`
	TokenDuration         int    `mapstructure:"TOKEN_DURATION"`
	CookieSecure          bool   `mapstructure:"COOKIE_SECURE"`
	CookieHTTPOnly        bool   `mapstructure:"COOKIE_HTTP_ONLY"`
	KeysPurpose           string `mapstructure:"KEYS_PURPOSE"`
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
