package ipc

import (
	"encoding/json"
	"net"
	"sync"

	"github.com/AvengeMedia/dankgo/log"
)

type ConnWriter struct {
	mu   sync.Mutex
	conn net.Conn
}

func NewConnWriter(conn net.Conn) *ConnWriter {
	return &ConnWriter{conn: conn}
}

func (w *ConnWriter) WriteResponse(resp any) error {
	data, err := json.Marshal(resp)
	if err != nil {
		log.Warnf("ipc encode response: %v", err)
		return err
	}
	return w.write(append(data, '\n'))
}

func (w *ConnWriter) WriteEvent(ev Event) error {
	envelope := map[string]any{
		"event": ev.Topic,
		"data":  ev.Data,
	}
	data, err := json.Marshal(envelope)
	if err != nil {
		log.Warnf("ipc encode event: %v", err)
		return err
	}
	return w.write(append(data, '\n'))
}

func (w *ConnWriter) write(data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if _, err := w.conn.Write(data); err != nil {
		log.Debugf("ipc write: %v", err)
		return err
	}
	return nil
}
