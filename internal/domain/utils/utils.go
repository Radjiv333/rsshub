package utils

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"RSSHub/pkg/config"
)

func ParseIntervalToDuration(intervalStr string) (time.Duration, error) {
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

	interval, err := ParseIntervalToDuration(envInterval)
	if err != nil {
		return 0, err
	}
	return interval, nil
}

func GetAndParseCliInterval() (time.Duration, error) {
	envInterval := config.GetEnvInterval()
	if len(envInterval) < 2 {
		return 0, fmt.Errorf("env value for cli_interval is invalid!")
	}

	interval, err := ParseIntervalToDuration(envInterval)
	if err != nil {
		return 0, err
	}
	return interval, nil
}

func ParseDurationToInterval(duration time.Duration) (string, error) {
	if duration <= 0 {
		return "", errors.New("duration must be greater than zero")
	}

	// Extract the number of seconds, minutes, hours, etc.
	if duration%time.Hour == 0 {
		// Duration is in hours
		hours := int(duration / time.Hour)
		return fmt.Sprintf("%dh", hours), nil
	} else if duration%time.Minute == 0 {
		// Duration is in minutes
		minutes := int(duration / time.Minute)
		return fmt.Sprintf("%dm", minutes), nil
	} else if duration%time.Second == 0 {
		// Duration is in seconds
		seconds := int(duration / time.Second)
		return fmt.Sprintf("%ds", seconds), nil
	} else {
		// Default fallback, returning in seconds if not perfectly divisible
		seconds := int(duration / time.Second)
		return fmt.Sprintf("%ds", seconds), nil
	}
}

func GetAndParseWorkersNum() (string, error) {
	workersNum := config.GetEnvWorkersNum()
	if _, err := strconv.Atoi(workersNum); err != nil {
		return "", err
	}
	return workersNum, nil
}
