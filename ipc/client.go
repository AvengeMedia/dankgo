package ipc

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AvengeMedia/dankgo/paths"
)

type Client struct {
	conn net.Conn
	r    *bufio.Reader
}

func Dial(socketPath string) (*Client, error) {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", socketPath, err)
	}

	r := bufio.NewReader(conn)
	if _, err := r.ReadBytes('\n'); err != nil {
		conn.Close()
		return nil, fmt.Errorf("read capabilities: %w", err)
	}

	return &Client{conn: conn, r: r}, nil
}

func (c *Client) Call(req Request) (*Response[any], error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	if _, err := c.conn.Write(append(data, '\n')); err != nil {
		return nil, err
	}
	line, err := c.r.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	var resp Response[any]
	if err := json.Unmarshal(line, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) Close() error { return c.conn.Close() }

func FindRunningSocket(appName string) (string, error) {
	app := paths.New(appName)
	dir := app.SocketDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	prefix := appName + "-"
	for _, entry := range entries {
		switch {
		case !strings.HasPrefix(entry.Name(), prefix):
			continue
		case !strings.HasSuffix(entry.Name(), ".sock"):
			continue
		}

		path := filepath.Join(dir, entry.Name())
		conn, err := net.DialTimeout("unix", path, 500*time.Millisecond)
		if err != nil {
			continue
		}
		conn.Close()
		return path, nil
	}
	return "", errors.New("no running " + appName + " socket found")
}
