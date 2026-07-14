package shellapp

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const pidFileExtension = ".pid"

func (a *App) pidFilePrefix() string { return a.cfg.ID + "-" }

func (a *App) pidFilePath() string {
	return filepath.Join(a.runtimeDir(), fmt.Sprintf("%s%d%s", a.pidFilePrefix(), os.Getpid(), pidFileExtension))
}

func (a *App) writePIDFile(childPID int) error {
	return os.WriteFile(a.pidFilePath(), []byte(strconv.Itoa(childPID)), 0o644)
}

func (a *App) removePIDFile() {
	os.Remove(a.pidFilePath())
}

func (a *App) removeAllPIDFiles() {
	entries, err := os.ReadDir(a.runtimeDir())
	if err != nil {
		return
	}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, a.pidFilePrefix()) && strings.HasSuffix(name, pidFileExtension) {
			os.Remove(filepath.Join(a.runtimeDir(), name))
		}
	}
}

// PIDs returns the live UI child PIDs and their supervising parents,
// reaping PID files whose processes are gone.
func (a *App) PIDs() []int {
	dir := a.runtimeDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var pids []int

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, a.pidFilePrefix()) || !strings.HasSuffix(name, pidFileExtension) {
			continue
		}

		pidFile := filepath.Join(dir, name)
		data, err := os.ReadFile(pidFile)
		if err != nil {
			continue
		}

		childPID, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err != nil {
			os.Remove(pidFile)
			continue
		}

		if !processAlive(childPID) {
			os.Remove(pidFile)
			continue
		}

		pids = append(pids, childPID)

		parentPIDStr := strings.TrimSuffix(strings.TrimPrefix(name, a.pidFilePrefix()), pidFileExtension)
		parentPID, err := strconv.Atoi(parentPIDStr)
		if err != nil {
			continue
		}
		if processAlive(parentPID) {
			pids = append(pids, parentPID)
		}
	}

	return pids
}

func processAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}
