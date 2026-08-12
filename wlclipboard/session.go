//go:build unix

// Package wlclipboard reads, watches, and owns the Wayland clipboard through
// the ext_data_control_v1 protocol without external binaries.
package wlclipboard

import (
	"fmt"

	wlclient "github.com/AvengeMedia/dankgo/wayland/client"
	"github.com/AvengeMedia/dankgo/wayland/ext_data_control"
)

type session struct {
	display        *wlclient.Display
	ctx            *wlclient.Context
	registry       *wlclient.Registry
	seat           *wlclient.Seat
	dataControlMgr *ext_data_control.ExtDataControlManagerV1
}

// connectSession opens a Wayland connection and binds the seat plus the
// data-control global if the compositor advertises it.
func connectSession() (*session, error) {
	display, err := wlclient.Connect("")
	if err != nil {
		return nil, fmt.Errorf("wayland connect: %w", err)
	}

	s := &session{display: display, ctx: display.Context()}

	registry, err := display.GetRegistry()
	if err != nil {
		display.Destroy()
		return nil, fmt.Errorf("get registry: %w", err)
	}
	s.registry = registry

	var bindErr error
	bind := func(name uint32, iface string, version uint32, proxy wlclient.Proxy) {
		if err := registry.Bind(name, iface, version, proxy); err != nil {
			bindErr = fmt.Errorf("bind %s: %w", iface, err)
		}
	}

	registry.SetGlobalHandler(func(e wlclient.RegistryGlobalEvent) {
		switch e.Interface {
		case ext_data_control.ExtDataControlManagerV1InterfaceName:
			mgr := ext_data_control.NewExtDataControlManagerV1(s.ctx)
			bind(e.Name, e.Interface, e.Version, mgr)
			s.dataControlMgr = mgr
		case "wl_seat":
			if s.seat != nil {
				return
			}
			seat := wlclient.NewSeat(s.ctx)
			bind(e.Name, e.Interface, e.Version, seat)
			s.seat = seat
		}
	})

	display.Roundtrip()
	display.Roundtrip()

	if bindErr != nil {
		s.Close()
		return nil, bindErr
	}
	return s, nil
}

func (s *session) requireDataControl() (*ext_data_control.ExtDataControlManagerV1, error) {
	switch {
	case s.dataControlMgr == nil:
		return nil, fmt.Errorf("compositor does not support ext_data_control_manager_v1")
	case s.seat == nil:
		return nil, fmt.Errorf("no seat available")
	default:
		return s.dataControlMgr, nil
	}
}

func (s *session) Close() {
	if s.dataControlMgr != nil {
		s.dataControlMgr.Destroy()
	}
	if s.registry != nil {
		s.registry.Destroy()
	}
	s.display.Destroy()
}

// Available reports whether the compositor supports the data-control
// protocol this package requires.
func Available() bool {
	s, err := connectSession()
	if err != nil {
		return false
	}
	defer s.Close()

	_, err = s.requireDataControl()
	return err == nil
}
