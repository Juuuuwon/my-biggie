package main

import (
	"go.uber.org/zap"

	"github.com/spf13/viper"
)

// initConfig loads configuration via Viper.
func initConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		logger.Warn("no config file found, using defaults", zap.Error(err))
	}
	// Default environment variables.
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("STARTUP_DELAY_SECOND", "0")
}

// processPort reads the PORT env variable and uses processRandomInt to support "RANDOM" values.
func processPort() int {
	portStr := viper.GetString("PORT")
	port, err := processRandomInt(portStr, 1024, 65535)
	if err != nil {
		logger.Warn("invalid PORT env var", zap.Error(err))
		return 8080
	}
	return port
}
