package config

import (
	"os"
)

func GetEnvInterval() string {
	envInterval := os.Getenv("CLI_APP_TIMER_INTERVAL")

	return envInterval
}
