package utils

import (
	"fmt"
	"strconv"
	"time"

	"RSSHub/pkg/config"
)

func ParseInterval(intervalStr string) (time.Duration, error) {
	if len(intervalStr) < 2 {
		return 0, fmt.Errorf("env value for db_interval is invalid!")
	}

	unit := intervalStr[len(intervalStr)-1]
	value := intervalStr[:len(intervalStr)-1]

	interval, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid interval value %q: %w", value, err)
	}

	switch unit {
	case 's':
		return time.Duration(interval) * time.Second, nil
	case 'm':
		return time.Duration(interval) * time.Minute, nil
	case 'h':
		return time.Duration(interval) * time.Hour, nil
	case 'd':
		return time.Duration(interval) * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unsupported unit: %c", unit)
	}
}

func GetAndParseDBInterval() (time.Duration, error) {
	envInterval := config.GetEnvDBInterval()

	interval, err := ParseInterval(envInterval)
	if err != nil {
		return 0, err
	}
	return interval, nil
}

func GetAndParseInterval() (time.Duration, error) {
	envInterval := config.GetEnvInterval()
	if len(envInterval) < 2 {
		return 0, fmt.Errorf("env value for cli_interval is invalid!")
	}

	interval, err := ParseInterval(envInterval)
	if err != nil {
		return 0, err
	}
	return interval, nil
}
