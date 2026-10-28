package notify

import (
	"fmt"
	"strings"

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

// Send emits a desktop notification with default normal urgency.
func (n *Notifier) Send(summary, body string) error {
	return n.SendWithUrgency(summary, body, "normal")
}

// SendWithUrgency emits a desktop notification with the specified urgency ("low", "normal", "critical").
func (n *Notifier) SendWithUrgency(summary, body, urgency string) error {
	conn, err := dbus.SessionBus()
	if err != nil {
		return fmt.Errorf("failed to connect to session bus: %w", err)
	}

	obj := conn.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")

	var urgencyVal byte = 1 // 0: low, 1: normal, 2: critical
	var expireTimeout int32 = 8000

	switch strings.ToLower(strings.TrimSpace(urgency)) {
	case "low":
		urgencyVal = 0
		expireTimeout = 5000
	case "critical":
		urgencyVal = 2
		expireTimeout = 0 // Critical alerts persist until dismissed by user
	default:
		urgencyVal = 1
		expireTimeout = 8000
	}

	// D-Bus org.freedesktop.Notifications.Notify signature:
	// (susssasa{sv}i) -> (u)
	hints := map[string]dbus.Variant{
		"urgency": dbus.MakeVariant(urgencyVal),
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
		expireTimeout,
	)

	if call.Err != nil {
		return fmt.Errorf("failed to send notification: %w", call.Err)
	}

	return nil
}
