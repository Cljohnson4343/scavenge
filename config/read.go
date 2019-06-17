package config

import (
	"github.com/spf13/viper"
)

// Read in a config file if provided; else, look for a config file and read
// it in
func Read(configFile string) error {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$GOPATH/src/github.com/cljohnson4343/scavenge")
		viper.AddConfigPath("/etc/scavenge")
		viper.AddConfigPath("$HOME/.scavenge")
	}

	return viper.ReadInConfig()
}
