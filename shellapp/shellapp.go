// Package shellapp runs a quickshell-based desktop app: it supervises the
// UI process, daemonizes, tracks instances via PID files, resolves the
// shell config dir, and talks to a running instance over dankgo/ipc.
package shellapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AvengeMedia/dankgo/paths"
	"github.com/spf13/cobra"
)

type Backend interface {
	SocketPath() string
	Close()
}

type Embedded interface {
	Available() bool
	Extract(baseDir string) (string, error)
	Prune(baseDir, keep string)
}

type Config struct {
	ID         string // xdg/socket/pidfile identity, e.g. "dankcal"
	EnvPrefix  string // default: upper(ID); drives <P>_SOCKET, <P>_SHELL_DIR, <P>_LOG_*
	QSAppID    string // e.g. "com.danklinux.dankcalendar"
	Version    string
	Embedded   Embedded
	Boot       func(ctx context.Context) (Backend, error)
	PreLaunch  func()
	ExtraEnv   func(configPath string) []string
	OnUIExit   func(exitCode int, uptime time.Duration, stderrTail string)
	ShowMethod string // IPC method `run` calls on a live instance instead of relaunching; empty relaunches
}

type App struct {
	cfg            Config
	app            paths.App
	customConfig   string
	configPath     string
	sessionManaged bool
	startHidden    bool
}

func New(cfg Config) *App {
	if cfg.EnvPrefix == "" {
		cfg.EnvPrefix = strings.ToUpper(cfg.ID)
	}
	return &App{cfg: cfg, app: paths.New(cfg.ID)}
}

func (a *App) CustomConfigVar() *string { return &a.customConfig }

func (a *App) ConfigPath() string { return a.configPath }

func (a *App) shellDirEnv() string { return a.cfg.EnvPrefix + "_SHELL_DIR" }

func (a *App) runtimeDir() string { return a.app.SocketDir() }

func (a *App) stateFile() string {
	return filepath.Join(a.runtimeDir(), a.cfg.ID+".path")
}

func binaryName() string { return filepath.Base(os.Args[0]) }

// ResolveConfig resolves the quickshell config dir. Explicit overrides win
// (custom path, then <PREFIX>_SHELL_DIR), then the path a running instance
// is using, then the UI embedded in the binary. There is no implicit
// filesystem search.
func (a *App) ResolveConfig(_ *cobra.Command, _ []string) error {
	if a.customConfig != "" {
		if err := validateShellDir(a.customConfig); err != nil {
			return fmt.Errorf("custom config path: %w", err)
		}
		a.configPath = a.customConfig
		return nil
	}

	if dir := os.Getenv(a.shellDirEnv()); dir != "" {
		if err := validateShellDir(dir); err != nil {
			return fmt.Errorf("%s: %w", a.shellDirEnv(), err)
		}
		a.configPath = dir
		return nil
	}

	if data, err := os.ReadFile(a.stateFile()); err == nil {
		switch len(a.PIDs()) {
		case 0:
			os.Remove(a.stateFile())
		default:
			statePath := strings.TrimSpace(string(data))
			if err := validateShellDir(statePath); err == nil {
				a.configPath = statePath
				return nil
			}
			os.Remove(a.stateFile())
		}
	}

	return a.useEmbeddedShell()
}

func validateShellDir(dir string) error {
	shellPath := filepath.Join(dir, "shell.qml")
	info, err := os.Stat(shellPath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("expected file, got directory: %s", shellPath)
	}
	return nil
}

func (a *App) useEmbeddedShell() error {
	if a.cfg.Embedded == nil || !a.cfg.Embedded.Available() {
		return fmt.Errorf("this %s build has no embedded UI; pass -c <dir> or set %s", binaryName(), a.shellDirEnv())
	}

	baseDir := filepath.Join(a.runtimeDir(), a.cfg.ID+"-shell")
	dir, err := a.cfg.Embedded.Extract(baseDir)
	if err != nil {
		return fmt.Errorf("extract embedded UI: %w", err)
	}

	if len(a.PIDs()) == 0 {
		a.cfg.Embedded.Prune(baseDir, dir)
	}

	a.configPath = dir
	return nil
}

func (a *App) writeConfigState() error {
	return os.WriteFile(a.stateFile(), []byte(a.configPath), 0o644)
}

func (a *App) clearConfigState() {
	os.Remove(a.stateFile())
}
