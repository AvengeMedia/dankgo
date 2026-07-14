package shellapp

import (
	"context"
	"os"
	"os/exec"

	"github.com/AvengeMedia/dankgo/log"
)

func hasSystemdRun() bool {
	_, err := exec.LookPath("systemd-run")
	return err == nil
}

func (a *App) appendLogEnv(env []string) []string {
	for _, key := range []string{a.cfg.EnvPrefix + "_LOG_LEVEL", a.cfg.EnvPrefix + "_LOG_FILE"} {
		if v := os.Getenv(key); v != "" {
			env = append(env, key+"="+v)
		}
	}
	if rules := log.GetQtLoggingRules(); rules != "" {
		env = append(env, "QT_LOGGING_RULES="+rules)
	}
	return env
}

func appendQtEnv(env []string) []string {
	if os.Getenv("QT_QPA_PLATFORMTHEME") == "" {
		env = append(env, "QT_QPA_PLATFORMTHEME=gtk3")
	}
	if os.Getenv("QT_QPA_PLATFORMTHEME_QT6") == "" {
		env = append(env, "QT_QPA_PLATFORMTHEME_QT6=gtk3")
	}
	if os.Getenv("QT_QPA_PLATFORM") == "" {
		env = append(env, "QT_QPA_PLATFORM=wayland;xcb")
	}
	return env
}

func (a *App) buildUICommand(ctx context.Context, socketPath string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, "qs", "-p", a.configPath)
	env := append(os.Environ(),
		a.cfg.EnvPrefix+"_SOCKET="+socketPath,
		"QS_APP_ID="+a.cfg.QSAppID,
	)

	if a.startHidden {
		env = append(env, a.cfg.EnvPrefix+"_START_HIDDEN=1")
	}

	if a.sessionManaged && hasSystemdRun() {
		env = append(env, a.cfg.EnvPrefix+"_DEFAULT_LAUNCH_PREFIX=systemd-run --user --scope")
	}

	env = appendQtEnv(env)
	env = a.appendLogEnv(env)
	cmd.Env = env
	return cmd
}
