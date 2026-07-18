package shellapp

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

const (
	pidFileExtension     = ".pid"
	sessionFileExtension = ".session"
)

func (a *App) pidFilePrefix() string { return a.cfg.ID + "-" }

func (a *App) pidFilePath() string {
	return filepath.Join(a.runtimeDir(), fmt.Sprintf("%s%d%s", a.pidFilePrefix(), os.Getpid(), pidFileExtension))
}

func (a *App) sessionFilePath() string {
	return filepath.Join(a.runtimeDir(), fmt.Sprintf("%s%d%s", a.pidFilePrefix(), os.Getpid(), sessionFileExtension))
}

func (a *App) writePIDFile(childPID int) error {
	if display := os.Getenv("WAYLAND_DISPLAY"); display != "" {
		if err := os.WriteFile(a.sessionFilePath(), []byte(display), 0o644); err != nil {
			return err
		}
	}
	return os.WriteFile(a.pidFilePath(), []byte(strconv.Itoa(childPID)), 0o644)
}

func (a *App) removePIDFile() {
	os.Remove(a.pidFilePath())
	os.Remove(a.sessionFilePath())
}

func (a *App) removeAllPIDFiles() {
	entries, err := os.ReadDir(a.runtimeDir())
	if err != nil {
		return
	}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, a.pidFilePrefix()) {
			continue
		}
		if strings.HasSuffix(name, pidFileExtension) || strings.HasSuffix(name, sessionFileExtension) {
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

func (a *App) sessionParentPID(display string) (int, bool) {
	if display == "" {
		return 0, false
	}

	dir := a.runtimeDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, false
	}

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, a.pidFilePrefix()) || !strings.HasSuffix(name, sessionFileExtension) {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || strings.TrimSpace(string(data)) != display {
			continue
		}

		parentStr := strings.TrimSuffix(strings.TrimPrefix(name, a.pidFilePrefix()), sessionFileExtension)
		parentPID, err := strconv.Atoi(parentStr)
		if err != nil {
			continue
		}

		return parentPID, true
	}

	return 0, false
}

func (a *App) firstChildPID() (int, bool) {
	entries, err := os.ReadDir(a.runtimeDir())
	if err != nil {
		return 0, false
	}

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, a.pidFilePrefix()) || !strings.HasSuffix(name, pidFileExtension) {
			continue
		}

		data, err := os.ReadFile(filepath.Join(a.runtimeDir(), name))
		if err != nil {
			continue
		}

		pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err != nil || !processAlive(pid) {
			continue
		}

		return pid, true
	}

	return 0, false
}

// SessionPID returns the UI child PID for the current WAYLAND_DISPLAY
// session, falling back to the first live instance.
func (a *App) SessionPID() (int, bool) {
	parentPID, ok := a.sessionParentPID(os.Getenv("WAYLAND_DISPLAY"))
	if !ok {
		return a.firstChildPID()
	}

	pidFile := filepath.Join(a.runtimeDir(), fmt.Sprintf("%s%d%s", a.pidFilePrefix(), parentPID, pidFileExtension))
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return a.firstChildPID()
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || !processAlive(pid) {
		return a.firstChildPID()
	}

	return pid, true
}

// SessionSocketPath returns the backend socket of the instance supervising
// the current WAYLAND_DISPLAY session.
func (a *App) SessionSocketPath() (string, bool) {
	parentPID, ok := a.sessionParentPID(os.Getenv("WAYLAND_DISPLAY"))
	if !ok {
		return "", false
	}

	socket := filepath.Join(a.runtimeDir(), fmt.Sprintf("%s%d.sock", a.pidFilePrefix(), parentPID))
	if _, err := os.Stat(socket); err != nil {
		return "", false
	}

	return socket, true
}
