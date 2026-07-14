package shellapp

import (
	"errors"
	"time"

	"github.com/AvengeMedia/dankgo/ipc"
)

const (
	launchDeadline    = 15 * time.Second
	launchRetryPeriod = 250 * time.Millisecond
)

func (a *App) CallUI(method string, params map[string]any) error {
	socketPath, err := ipc.FindRunningSocket(a.cfg.ID)
	if err != nil {
		return err
	}

	client, err := ipc.Dial(socketPath)
	if err != nil {
		return err
	}
	defer client.Close()

	resp, err := client.Call(ipc.Request{ID: 1, Method: method, Params: params})
	if err != nil {
		return err
	}
	if resp.Error != "" {
		return errors.New(resp.Error)
	}
	return nil
}

// CallOrLaunch sends method to a running instance, cold-starting the
// daemon when none answers. With params it retries until the fresh
// instance's socket is up, so the request survives the cold start.
func (a *App) CallOrLaunch(method string, params map[string]any) error {
	if err := a.CallUI(method, params); err == nil {
		return nil
	}
	if err := a.RunDaemon(false); err != nil {
		return err
	}
	if len(params) == 0 {
		return nil
	}

	deadline := time.Now().Add(launchDeadline)
	for {
		if err := a.CallUI(method, params); err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return errors.New("timed out waiting for " + binaryName() + " to start")
		}
		time.Sleep(launchRetryPeriod)
	}
}
