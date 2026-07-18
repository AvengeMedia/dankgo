package ipc_test

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/AvengeMedia/dankgo/ipc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func startServer(t *testing.T, cfg ipc.Config, handler ipc.Handler) *ipc.Server {
	t.Helper()
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())

	srv := ipc.NewServer(cfg, handler)
	require.NoError(t, srv.Listen())

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() { _ = srv.Serve(ctx) }()
	t.Cleanup(func() { _ = srv.Close() })
	return srv
}

func TestHandshakeAndPing(t *testing.T) {
	srv := startServer(t, ipc.Config{
		AppName:      "danktest",
		APIVersion:   1,
		Capabilities: []string{"things", "subscribe"},
	}, nil)

	conn, err := net.Dial("unix", srv.SocketPath())
	require.NoError(t, err)
	defer conn.Close()

	r := bufio.NewReader(conn)
	line, err := r.ReadBytes('\n')
	require.NoError(t, err)
	assert.JSONEq(t, `{"apiVersion":1,"capabilities":["things","subscribe"]}`, string(line))

	_, err = conn.Write([]byte(`{"id":1,"method":"ping"}` + "\n"))
	require.NoError(t, err)

	line, err = r.ReadBytes('\n')
	require.NoError(t, err)
	assert.JSONEq(t, `{"id":1,"result":{"pong":true}}`, string(line))
}

func TestClientCallCustomHandler(t *testing.T) {
	handler := func(ctx context.Context, w *ipc.ConnWriter, req ipc.Request, sub *ipc.Subscriber) {
		switch req.Method {
		case "echo":
			ipc.Respond(w, req.ID, req.Params)
		default:
			ipc.RespondError(w, req.ID, "unknown method: "+req.Method)
		}
	}
	srv := startServer(t, ipc.Config{AppName: "danktest", APIVersion: 1}, handler)

	client, err := ipc.Dial(srv.SocketPath())
	require.NoError(t, err)
	defer client.Close()

	resp, err := client.Call(ipc.Request{ID: 7, Method: "echo", Params: map[string]any{"hello": "world"}})
	require.NoError(t, err)
	assert.Equal(t, 7, resp.ID)
	assert.Empty(t, resp.Error)
	require.NotNil(t, resp.Result)
	assert.Equal(t, map[string]any{"hello": "world"}, *resp.Result)

	resp, err = client.Call(ipc.Request{ID: 8, Method: "nope"})
	require.NoError(t, err)
	assert.Equal(t, "unknown method: nope", resp.Error)
}

func TestSubscribeAndPublish(t *testing.T) {
	subscribed := make(chan []string, 1)
	srv := startServer(t, ipc.Config{
		AppName:                "danktest",
		APIVersion:             1,
		DefaultSubscribeTopics: []string{"things"},
		OnSubscribe:            func(topics []string, sub *ipc.Subscriber) { subscribed <- topics },
	}, nil)

	conn, err := net.Dial("unix", srv.SocketPath())
	require.NoError(t, err)
	defer conn.Close()

	r := bufio.NewReader(conn)
	_, err = r.ReadBytes('\n')
	require.NoError(t, err)

	_, err = conn.Write([]byte(`{"id":1,"method":"subscribe"}` + "\n"))
	require.NoError(t, err)

	line, err := r.ReadBytes('\n')
	require.NoError(t, err)
	assert.JSONEq(t, `{"id":1,"result":{"topics":["things"]}}`, string(line))

	select {
	case topics := <-subscribed:
		assert.Equal(t, []string{"things"}, topics)
	case <-time.After(time.Second):
		t.Fatal("OnSubscribe not called")
	}

	require.Eventually(t, func() bool { return srv.Bus().HasSubscriber("things") }, time.Second, 10*time.Millisecond)
	srv.Bus().Publish("things", map[string]any{"n": 1})
	srv.Bus().Publish("other", map[string]any{"n": 2})

	line, err = r.ReadBytes('\n')
	require.NoError(t, err)
	var ev map[string]any
	require.NoError(t, json.Unmarshal(line, &ev))
	assert.Equal(t, "things", ev["event"])
	assert.Equal(t, map[string]any{"n": float64(1)}, ev["data"])
}

func TestSubscribeHandlerOverride(t *testing.T) {
	srv := startServer(t, ipc.Config{
		AppName:    "danktest",
		APIVersion: 1,
		SubscribeHandler: func(ctx context.Context, w *ipc.ConnWriter, req ipc.Request, sub *ipc.Subscriber) {
			ipc.Respond(w, req.ID, map[string]any{"method": req.Method, "custom": true})
		},
	}, nil)

	client, err := ipc.Dial(srv.SocketPath())
	require.NoError(t, err)
	defer client.Close()

	resp, err := client.Call(ipc.Request{ID: 1, Method: "subscribe", Params: map[string]any{"services": []any{"network"}}})
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"method": "subscribe", "custom": true}, *resp.Result)

	resp, err = client.Call(ipc.Request{ID: 2, Method: "unsubscribe"})
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"method": "unsubscribe", "custom": true}, *resp.Result)
}

func TestFindRunningSocket(t *testing.T) {
	srv := startServer(t, ipc.Config{AppName: "danktest", APIVersion: 1}, nil)

	path, err := ipc.FindRunningSocket("danktest")
	require.NoError(t, err)
	assert.Equal(t, srv.SocketPath(), path)

	_, err = ipc.FindRunningSocket("nosuchapp")
	assert.Error(t, err)
}

func TestMux(t *testing.T) {
	mux := ipc.NewMux()
	mux.Handle("version", func(ctx context.Context, w *ipc.ConnWriter, req ipc.Request, sub *ipc.Subscriber) {
		ipc.Respond(w, req.ID, map[string]any{"version": "1.0"})
	})
	mux.HandlePrefix("things.", func(ctx context.Context, w *ipc.ConnWriter, req ipc.Request, sub *ipc.Subscriber) {
		ipc.Respond(w, req.ID, map[string]any{"method": req.Method})
	})
	srv := startServer(t, ipc.Config{AppName: "danktest", APIVersion: 1}, mux.ServeIPC)

	client, err := ipc.Dial(srv.SocketPath())
	require.NoError(t, err)
	defer client.Close()

	resp, err := client.Call(ipc.Request{ID: 1, Method: "version"})
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"version": "1.0"}, *resp.Result)

	resp, err = client.Call(ipc.Request{ID: 2, Method: "things.list"})
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"method": "things.list"}, *resp.Result)

	resp, err = client.Call(ipc.Request{ID: 3, Method: "bogus"})
	require.NoError(t, err)
	assert.Equal(t, "unknown method: bogus", resp.Error)
}
