package notify

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

// Notifier sends desktop notifications via org.freedesktop.Notifications over D-Bus.
type Notifier struct {
	appName string
}

// New creates a new Notifier with the given application name.
func New(appName string) *Notifier {
	return &Notifier{
		appName: appName,
	}
}

// Send emits a desktop notification.
func (n *Notifier) Send(summary, body string) error {
	conn, err := dbus.SessionBus()
	if err != nil {
		return fmt.Errorf("failed to connect to session bus: %w", err)
	}

	obj := conn.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")

	// D-Bus org.freedesktop.Notifications.Notify signature:
	// (susssasa{sv}i) -> (u)
	hints := map[string]dbus.Variant{
		"urgency": dbus.MakeVariant(byte(1)), // 0: low, 1: normal, 2: critical
	}

	call := obj.Call(
		"org.freedesktop.Notifications.Notify",
		0,
		n.appName,
		uint32(0),
		"dialog-warning",
		summary,
		body,
		[]string{},
		hints,
		int32(8000), // 8 seconds
	)

	if call.Err != nil {
		return fmt.Errorf("failed to send notification: %w", call.Err)
	}

	return nil
}
