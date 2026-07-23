package portal

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/godbus/dbus/v5"

	"github.com/AvengeMedia/dankgo/dbusutil"
)

const secretInterface = "org.freedesktop.portal.Secret"

type SecretOptions struct {
	// Token is the opaque value returned by a previous retrieval; optional.
	Token string
}

type SecretResult struct {
	Secret []byte
	Token  string
}

// RetrieveSecret asks the Secret portal for the application's master secret,
// delivered through a pipe handed to the portal. Sandboxed callers receive a
// stable per-app-ID secret suitable for encrypting local storage.
func RetrieveSecret(ctx context.Context, conn *dbus.Conn, opts SecretOptions) (SecretResult, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return SecretResult{}, fmt.Errorf("create secret pipe: %w", err)
	}
	defer r.Close()

	token := newHandleToken()
	waiter, err := newResponseWaiter(conn, requestPath(conn, token))
	if err != nil {
		w.Close()
		return SecretResult{}, err
	}
	defer waiter.close()

	options := map[string]dbus.Variant{"handle_token": dbus.MakeVariant(token)}
	if opts.Token != "" {
		options["token"] = dbus.MakeVariant(opts.Token)
	}

	obj := conn.Object(busName, objectPath)
	call := obj.Call(secretInterface+".RetrieveSecret", 0, dbus.UnixFD(w.Fd()), options)
	// The portal holds its own duplicate; closing our write end lets the
	// reader observe EOF once the portal finishes writing.
	w.Close()
	if call.Err != nil {
		return SecretResult{}, fmt.Errorf("retrieve portal secret: %w", call.Err)
	}

	var handle dbus.ObjectPath
	if err := call.Store(&handle); err != nil {
		return SecretResult{}, fmt.Errorf("read secret request handle: %w", err)
	}
	if err := waiter.redirect(handle); err != nil {
		return SecretResult{}, err
	}

	type readResult struct {
		data []byte
		err  error
	}
	read := make(chan readResult, 1)
	go func() {
		data, err := io.ReadAll(r)
		read <- readResult{data, err}
	}()

	results, err := waiter.wait(ctx)
	if err != nil {
		return SecretResult{}, err
	}

	res := <-read
	if res.err != nil {
		return SecretResult{}, fmt.Errorf("read portal secret: %w", res.err)
	}
	return SecretResult{
		Secret: res.data,
		Token:  dbusutil.GetOr(results, "token", ""),
	}, nil
}
