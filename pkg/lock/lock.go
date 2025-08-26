package lock

import (
	"fmt"
	"os"
)

const lockFile = "/tmp/rsshub.lock"

func Acquire() error {
	if _, err := os.Stat(lockFile); err == nil {
		return fmt.Errorf("fetch command already running (lock file exists)")
	}
	pid := os.Getpid()
	return os.WriteFile(lockFile, []byte(fmt.Sprintf("%d", pid)), 0o644)
}

func Release() {
	_ = os.Remove(lockFile)
}
