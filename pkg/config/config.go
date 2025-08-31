package config

import (
	"RSSHub/pkg/logger"
	"os"
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

func GetEnvWorkersNum() string {
	workers := os.Getenv("CLI_APP_WORKERS_COUNT")
	logger.Debug("Getting env value of workers", "workers", workers)
	return workers
}
