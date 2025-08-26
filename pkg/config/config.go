package config

import (
	"os"

	"RSSHub/pkg/logger"
)

func GetEnvInterval() string {
	envInterval := os.Getenv("CLI_APP_TIMER_INTERVAL")
	logger.Info("Getting env value of timer_interval", "timer_interval", envInterval)
	return envInterval
}
