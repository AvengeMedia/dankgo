package shellapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/AvengeMedia/dankgo/ipc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testApp(t *testing.T) *App {
	t.Helper()
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	return New(Config{ID: "danktest", QSAppID: "com.danklinux.danktest", Version: "0.0.1"})
}

func writeShellDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "shell.qml"), []byte("import QtQuick\n"), 0o644))
	return dir
}

func TestEnvPrefixDefaultsToUpperID(t *testing.T) {
	a := New(Config{ID: "danktest"})
	assert.Equal(t, "DANKTEST", a.cfg.EnvPrefix)

	a = New(Config{ID: "danktest", EnvPrefix: "CUSTOM"})
	assert.Equal(t, "CUSTOM", a.cfg.EnvPrefix)
}

func TestPIDsReapsDeadEntries(t *testing.T) {
	a := testApp(t)

	require.NoError(t, a.writePIDFile(os.Getpid()))
	deadFile := filepath.Join(a.runtimeDir(), fmt.Sprintf("danktest-%d.pid", 999999))
	require.NoError(t, os.WriteFile(deadFile, []byte("999998"), 0o644))
	junkFile := filepath.Join(a.runtimeDir(), "danktest-junk.pid")
	require.NoError(t, os.WriteFile(junkFile, []byte("not-a-pid"), 0o644))

	pids := a.PIDs()
	assert.Contains(t, pids, os.Getpid())
	assert.NotContains(t, pids, 999998)

	_, err := os.Stat(deadFile)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(junkFile)
	assert.True(t, os.IsNotExist(err))
}

func TestPIDsIncludesLiveParentFromFilename(t *testing.T) {
	a := testApp(t)

	parent := os.Getppid()
	file := filepath.Join(a.runtimeDir(), fmt.Sprintf("danktest-%d.pid", parent))
	require.NoError(t, os.WriteFile(file, []byte(strconv.Itoa(os.Getpid())), 0o644))

	pids := a.PIDs()
	assert.Contains(t, pids, os.Getpid())
	assert.Contains(t, pids, parent)
}

func TestResolveConfigCustomPathWins(t *testing.T) {
	a := testApp(t)
	dir := writeShellDir(t)
	t.Setenv("DANKTEST_SHELL_DIR", writeShellDir(t))

	*a.CustomConfigVar() = dir
	require.NoError(t, a.ResolveConfig(nil, nil))
	assert.Equal(t, dir, a.ConfigPath())
}

func TestResolveConfigCustomPathValidated(t *testing.T) {
	a := testApp(t)

	*a.CustomConfigVar() = t.TempDir()
	err := a.ResolveConfig(nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "custom config path")
}

func TestResolveConfigEnvVar(t *testing.T) {
	a := testApp(t)
	dir := writeShellDir(t)
	t.Setenv("DANKTEST_SHELL_DIR", dir)

	require.NoError(t, a.ResolveConfig(nil, nil))
	assert.Equal(t, dir, a.ConfigPath())
}

func TestResolveConfigStateFileUsedOnlyWithLiveInstance(t *testing.T) {
	a := testApp(t)
	dir := writeShellDir(t)
	require.NoError(t, os.WriteFile(a.stateFile(), []byte(dir+"\n"), 0o644))

	err := a.ResolveConfig(nil, nil)
	require.Error(t, err, "no live instance and no embedded UI")
	_, statErr := os.Stat(a.stateFile())
	assert.True(t, os.IsNotExist(statErr), "stale state file should be removed")

	require.NoError(t, os.WriteFile(a.stateFile(), []byte(dir+"\n"), 0o644))
	require.NoError(t, a.writePIDFile(os.Getpid()))
	require.NoError(t, a.ResolveConfig(nil, nil))
	assert.Equal(t, dir, a.ConfigPath())
}

type fakeEmbedded struct {
	dir string
}

func (f fakeEmbedded) Available() bool { return f.dir != "" }

func (f fakeEmbedded) Extract(string) (string, error) { return f.dir, nil }

func (f fakeEmbedded) Prune(string, string) {}

func TestResolveConfigFallsBackToEmbedded(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	dir := writeShellDir(t)
	a := New(Config{ID: "danktest", Embedded: fakeEmbedded{dir: dir}})

	require.NoError(t, a.ResolveConfig(nil, nil))
	assert.Equal(t, dir, a.ConfigPath())
}

func TestResolveConfigErrorsWithoutEmbedded(t *testing.T) {
	a := testApp(t)

	err := a.ResolveConfig(nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no embedded UI")
}

func TestBuildUICommandEnv(t *testing.T) {
	a := testApp(t)
	a.configPath = "/some/shell"
	a.startHidden = true
	t.Setenv("QT_QPA_PLATFORM", "")
	os.Unsetenv("QT_QPA_PLATFORM")
	t.Setenv("DANKTEST_LOG_LEVEL", "debug")

	cmd := a.buildUICommand(context.Background(), "/run/danktest-1.sock")

	assert.Equal(t, []string{"qs", "-p", "/some/shell"}, cmd.Args)
	env := strings.Join(cmd.Env, "\n")
	assert.Contains(t, env, "DANKTEST_SOCKET=/run/danktest-1.sock")
	assert.Contains(t, env, "QS_APP_ID=com.danklinux.danktest")
	assert.Contains(t, env, "DANKTEST_START_HIDDEN=1")
	assert.Contains(t, env, "QT_QPA_PLATFORM=wayland;xcb")
	assert.Contains(t, env, "DANKTEST_LOG_LEVEL=debug")
}

func TestCallUI(t *testing.T) {
	a := testApp(t)

	srv := ipc.NewServer(ipc.Config{AppName: "danktest", APIVersion: 1},
		func(ctx context.Context, w *ipc.ConnWriter, req ipc.Request, sub *ipc.Subscriber) {
			switch req.Method {
			case "ui.show":
				ipc.Respond(w, req.ID, map[string]any{"ok": true})
			default:
				ipc.RespondError(w, req.ID, "unknown method: "+req.Method)
			}
		})
	require.NoError(t, srv.Listen())
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() { _ = srv.Serve(ctx) }()
	t.Cleanup(func() { _ = srv.Close() })

	assert.NoError(t, a.CallUI("ui.show", nil))
	assert.ErrorContains(t, a.CallUI("ui.bogus", nil), "unknown method")
}

func TestCallUIFailsWithoutInstance(t *testing.T) {
	a := testApp(t)
	assert.Error(t, a.CallUI("ui.show", nil))
}

func TestSessionFileLifecycle(t *testing.T) {
	a := testApp(t)
	t.Setenv("WAYLAND_DISPLAY", "wayland-7")

	require.NoError(t, a.writePIDFile(os.Getpid()))
	data, err := os.ReadFile(a.sessionFilePath())
	require.NoError(t, err)
	assert.Equal(t, "wayland-7", string(data))

	a.removePIDFile()
	_, err = os.Stat(a.sessionFilePath())
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(a.pidFilePath())
	assert.True(t, os.IsNotExist(err))
}

func TestSessionPIDPrefersMatchingDisplay(t *testing.T) {
	a := testApp(t)
	t.Setenv("WAYLAND_DISPLAY", "wayland-7")
	require.NoError(t, a.writePIDFile(os.Getpid()))

	otherPID := filepath.Join(a.runtimeDir(), fmt.Sprintf("danktest-%d.pid", 999999))
	require.NoError(t, os.WriteFile(otherPID, []byte(strconv.Itoa(os.Getppid())), 0o644))
	otherSession := filepath.Join(a.runtimeDir(), fmt.Sprintf("danktest-%d.session", 999999))
	require.NoError(t, os.WriteFile(otherSession, []byte("wayland-0"), 0o644))

	pid, ok := a.SessionPID()
	require.True(t, ok)
	assert.Equal(t, os.Getpid(), pid)
}

func TestSessionPIDFallsBackToFirstLiveInstance(t *testing.T) {
	a := testApp(t)
	t.Setenv("WAYLAND_DISPLAY", "wayland-7")

	file := filepath.Join(a.runtimeDir(), fmt.Sprintf("danktest-%d.pid", os.Getppid()))
	require.NoError(t, os.WriteFile(file, []byte(strconv.Itoa(os.Getpid())), 0o644))

	pid, ok := a.SessionPID()
	require.True(t, ok)
	assert.Equal(t, os.Getpid(), pid)
}

func TestSessionSocketPath(t *testing.T) {
	a := testApp(t)
	t.Setenv("WAYLAND_DISPLAY", "wayland-7")
	require.NoError(t, a.writePIDFile(os.Getpid()))

	_, ok := a.SessionSocketPath()
	assert.False(t, ok)

	socket := filepath.Join(a.runtimeDir(), fmt.Sprintf("danktest-%d.sock", os.Getpid()))
	require.NoError(t, os.WriteFile(socket, nil, 0o600))

	path, ok := a.SessionSocketPath()
	require.True(t, ok)
	assert.Equal(t, socket, path)
}

func TestHotReloadDisabledOutsideHome(t *testing.T) {
	a := testApp(t)
	a.configPath = filepath.Join(t.TempDir(), "shell")

	cmd := a.buildUICommand(context.Background(), "/run/danktest-1.sock")
	assert.Contains(t, strings.Join(cmd.Env, "\n"), "DANKTEST_DISABLE_HOT_RELOAD=1")

	home, err := os.UserHomeDir()
	require.NoError(t, err)
	a.configPath = filepath.Join(home, "shell")
	cmd = a.buildUICommand(context.Background(), "/run/danktest-1.sock")
	assert.NotContains(t, strings.Join(cmd.Env, "\n"), "DANKTEST_DISABLE_HOT_RELOAD")

	t.Setenv("DANKTEST_DISABLE_HOT_RELOAD", "0")
	a.configPath = filepath.Join(t.TempDir(), "shell")
	cmd = a.buildUICommand(context.Background(), "/run/danktest-1.sock")
	assert.NotContains(t, strings.Join(cmd.Env, "\n"), "DANKTEST_DISABLE_HOT_RELOAD=1")
}

func TestExtraEnvAppended(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	a := New(Config{ID: "danktest", ExtraEnv: func(configPath string) []string {
		return []string{"DANKTEST_EXECUTABLE=/usr/bin/danktest", "CONFIG=" + configPath}
	}})
	a.configPath = "/some/shell"

	env := strings.Join(a.buildUICommand(context.Background(), "/run/danktest-1.sock").Env, "\n")
	assert.Contains(t, env, "DANKTEST_EXECUTABLE=/usr/bin/danktest")
	assert.Contains(t, env, "CONFIG=/some/shell")
}
