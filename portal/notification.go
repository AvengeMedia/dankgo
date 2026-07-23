package portal

import (
	"fmt"

	"github.com/godbus/dbus/v5"

	"github.com/AvengeMedia/dankgo/dbusutil"
)

const notificationInterface = "org.freedesktop.portal.Notification"

const (
	PriorityLow    = "low"
	PriorityNormal = "normal"
	PriorityHigh   = "high"
	PriorityUrgent = "urgent"

	SoundDefault = "default"
	SoundSilence = "silence"
)

type Button struct {
	Label  string
	Action string
}

type Notification struct {
	Title string
	Body  string
	// Priority is one of the Priority constants; empty lets the server decide.
	Priority string
	// DefaultAction is invoked when the user activates the notification body.
	DefaultAction string
	Buttons       []Button
	// Sound is SoundDefault or SoundSilence. It requires interface version 2
	// and is dropped on older portals.
	Sound string
}

func (n Notification) data(version uint32) map[string]dbus.Variant {
	props := map[string]dbus.Variant{"title": dbus.MakeVariant(n.Title)}
	if n.Body != "" {
		props["body"] = dbus.MakeVariant(n.Body)
	}
	if n.Priority != "" {
		props["priority"] = dbus.MakeVariant(n.Priority)
	}
	if n.DefaultAction != "" {
		props["default-action"] = dbus.MakeVariant(n.DefaultAction)
	}
	if len(n.Buttons) > 0 {
		buttons := make([]map[string]dbus.Variant, 0, len(n.Buttons))
		for _, b := range n.Buttons {
			buttons = append(buttons, map[string]dbus.Variant{
				"label":  dbus.MakeVariant(b.Label),
				"action": dbus.MakeVariant(b.Action),
			})
		}
		props["buttons"] = dbus.MakeVariant(buttons)
	}
	if n.Sound != "" && version >= 2 {
		props["sound"] = dbus.MakeVariant(n.Sound)
	}
	return props
}

// NotificationClient sends notifications through the desktop portal and
// dispatches ActionInvoked signals. The caller owns the bus connection;
// Close only unsubscribes.
type NotificationClient struct {
	conn     *dbus.Conn
	version  uint32
	signals  chan *dbus.Signal
	onAction func(id, action string)
}

func NewNotificationClient(conn *dbus.Conn) (*NotificationClient, error) {
	if err := conn.AddMatchSignal(
		dbus.WithMatchInterface(notificationInterface),
		dbus.WithMatchObjectPath(objectPath),
	); err != nil {
		return nil, fmt.Errorf("subscribe portal notification signals: %w", err)
	}

	c := &NotificationClient{conn: conn, version: 1}
	if v, err := conn.Object(busName, objectPath).GetProperty(notificationInterface + ".version"); err == nil {
		c.version = dbusutil.AsOr(v, uint32(1))
	}

	c.signals = make(chan *dbus.Signal, 16)
	conn.Signal(c.signals)
	go c.dispatch()
	return c, nil
}

func (c *NotificationClient) Version() uint32 { return c.version }

// SetActionHandler must be called before the first Add. The handler runs on
// the signal goroutine, one signal at a time.
func (c *NotificationClient) SetActionHandler(fn func(id, action string)) {
	c.onAction = fn
}

// Add shows or replaces the notification with the given app-chosen id.
func (c *NotificationClient) Add(id string, n Notification) error {
	obj := c.conn.Object(busName, objectPath)
	call := obj.Call(notificationInterface+".AddNotification", 0, id, n.data(c.version))
	if call.Err != nil {
		return fmt.Errorf("add portal notification: %w", call.Err)
	}
	return nil
}

func (c *NotificationClient) Remove(id string) error {
	obj := c.conn.Object(busName, objectPath)
	call := obj.Call(notificationInterface+".RemoveNotification", 0, id)
	if call.Err != nil {
		return fmt.Errorf("remove portal notification: %w", call.Err)
	}
	return nil
}

func (c *NotificationClient) Close() {
	_ = c.conn.RemoveMatchSignal(
		dbus.WithMatchInterface(notificationInterface),
		dbus.WithMatchObjectPath(objectPath),
	)
	c.conn.RemoveSignal(c.signals)
	close(c.signals)
}

func (c *NotificationClient) dispatch() {
	for sig := range c.signals {
		if sig == nil || sig.Name != notificationInterface+".ActionInvoked" || len(sig.Body) < 2 {
			continue
		}
		if c.onAction == nil {
			continue
		}
		id, idOK := sig.Body[0].(string)
		action, actionOK := sig.Body[1].(string)
		if !idOK || !actionOK {
			continue
		}
		c.onAction(id, action)
	}
}
