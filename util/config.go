package util

import (
	"github.com/spf13/viper"
)

// Config - Wrapper for the toml config
type Config struct {
	NginXVersion   string
	PCREVersion    string
	ZLibVersion    string
	OpenSSLVersion string
	Configuration  string
	Modules        string
}

// GetConfig from the toml config
func GetConfig() (*Config, error) {
	// read nginx_params config
	viper.SetConfigType("toml")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()

	if err != nil {
		return nil, err
	}

	return &Config{
		viper.GetString("nginx_version"),
		viper.GetString("pcre_version"),
		viper.GetString("zlib_version"),
		viper.GetString("openssl_version"),
		viper.GetString("nginx_configuration"),
		viper.GetString("nginx_modules"),
	}, nil
}
