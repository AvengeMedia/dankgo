package portal

import (
	"testing"

	"github.com/godbus/dbus/v5"
	"github.com/stretchr/testify/assert"
)

func TestNotificationData(t *testing.T) {
	t.Run("full", func(t *testing.T) {
		n := Notification{
			Title:         "Reminder",
			Body:          "Standup in 5 minutes",
			Priority:      PriorityUrgent,
			DefaultAction: "default",
			Buttons: []Button{
				{Label: "Snooze", Action: "snooze"},
				{Label: "Dismiss", Action: "dismiss"},
			},
			Sound: SoundDefault,
		}

		props := n.data(2)
		assert.Equal(t, dbus.MakeVariant("Reminder"), props["title"])
		assert.Equal(t, dbus.MakeVariant("Standup in 5 minutes"), props["body"])
		assert.Equal(t, dbus.MakeVariant(PriorityUrgent), props["priority"])
		assert.Equal(t, dbus.MakeVariant("default"), props["default-action"])
		assert.Equal(t, dbus.MakeVariant(SoundDefault), props["sound"])

		buttons, ok := props["buttons"].Value().([]map[string]dbus.Variant)
		assert.True(t, ok)
		assert.Len(t, buttons, 2)
		assert.Equal(t, dbus.MakeVariant("Snooze"), buttons[0]["label"])
		assert.Equal(t, dbus.MakeVariant("snooze"), buttons[0]["action"])
	})

	t.Run("empty fields omitted", func(t *testing.T) {
		props := Notification{Title: "Hi"}.data(2)
		assert.Equal(t, map[string]dbus.Variant{"title": dbus.MakeVariant("Hi")}, props)
	})

	t.Run("sound dropped below version 2", func(t *testing.T) {
		props := Notification{Title: "Hi", Sound: SoundSilence}.data(1)
		_, ok := props["sound"]
		assert.False(t, ok)
	})
}
