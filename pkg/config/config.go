package config

import (
	"os"

	"RSSHub/pkg/logger"
)

func GetEnvInterval() string {
	envInterval := os.Getenv("CLI_APP_TIMER_INTERVAL")
	logger.Debug("Getting env value of cli_timer_interval", "cli_timer_interval", envInterval)
	return envInterval
}

func GetEnvDBInterval() string {
	envInterval := os.Getenv("DB_TIMER_INTERVAL")
	logger.Debug("Getting env value of db_timer_interval", "db_timer_interval", envInterval)
	return envInterval
}
