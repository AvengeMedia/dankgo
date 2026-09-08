package log

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const childEnv = "DANKGO_LOG_TEST_CHILD"

func TestLogFileEnvAppliesOnFirstInit(t *testing.T) {
	if os.Getenv(childEnv) != "" {
		ApplyEnvOverrides()
		Infof("child line \x1b[31mred\x1b[0m")
		return
	}

	logPath := filepath.Join(t.TempDir(), "child.log")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^"+t.Name()+"$")
	cmd.Env = append(os.Environ(), childEnv+"=1", envPrefix+"_LOG_FILE="+logPath)

	out, err := cmd.CombinedOutput()
	require.NoError(t, ctx.Err(), "child deadlocked with %s_LOG_FILE set", envPrefix)
	require.NoError(t, err, "child failed: %s", out)

	data, err := os.ReadFile(logPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "child line")
	assert.NotContains(t, string(data), "\x1b[")
}

func TestSetLogFileTeesStderrAndDetaches(t *testing.T) {
	var stderr bytes.Buffer
	original := logStderr
	logStderr = &stderr
	t.Cleanup(func() {
		logStderr = original
		_ = SetLogFile("")
	})

	logPath := filepath.Join(t.TempDir(), "app.log")
	require.NoError(t, SetLogFile(logPath))

	Infof("attached")

	data, err := os.ReadFile(logPath)
	require.NoError(t, err)
	assert.Contains(t, string(data), "attached")
	assert.Contains(t, stderr.String(), "attached")

	require.NoError(t, SetLogFile(""))
	Infof("detached")

	data, err = os.ReadFile(logPath)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "detached")
	assert.Contains(t, stderr.String(), "detached")
}

func TestSetLogFileRejectsUnwritablePath(t *testing.T) {
	err := SetLogFile(filepath.Join(t.TempDir(), "missing", "app.log"))
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "no such file or directory"), err.Error())
}
