package config

import "github.com/spf13/viper"

type Config struct {
	DbHost       string `mapstructure:"DB_HOST"`
	DbUser       string `mapstructure:"DB_USER"`
	DbPassword   string `mapstructure:"DB_PASSWORD"`
	DbName       string `mapstructure:"DB_NAME"`
	DbPort       string `mapstructure:"DB_PORT"`
	Port         string `mapstructure:"PORT"`
	SSLmode      string `mapstructure:"SSL_MODE"`
	JwtSecret    string `mapstructure:"JWT_SECRET_KEY"`
	AutoComplete string `mapstructure:"AUTO_COMPLETE_MINUTES"`
	RateLimiter  string `mapstructure:"RATE_LIMIT"`
}

func LoadConfig() (*Config, error) {
	var config Config

	viper.AddConfigPath("./")
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
