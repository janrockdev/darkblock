package util

import (
	"errors"
	"path/filepath"
	"runtime"

	"github.com/janrockdev/darkblock/config"
	"github.com/spf13/viper"
)

// LoadConfig reads the config file and returns a ConfigFile struct
func LoadConfig(path ...string) *config.ConfigFile {
	_, b, _, _ := runtime.Caller(0)
	basePath := filepath.Dir(b)
	rootPath := filepath.Join(basePath, "..")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath(rootPath)
	if err := viper.ReadInConfig(); err != nil {
		panic("failed to read config file" + err.Error())
	}

	var config config.ConfigFile
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			Logger.Error().Msgf("error reading config file: %v", err)
			return nil
		}
	}

	if err := viper.Unmarshal(&config); err != nil {
		Logger.Error().Msgf("unable to decode into struct: %v", err)
		return nil
	}

	return &config
}
