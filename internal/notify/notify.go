package notify

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
)

// Notifier sends desktop notifications via org.freedesktop.Notifications over D-Bus.
type Notifier struct {
	appName string
	sem     chan struct{}
}

// New creates a new Notifier with the given application name.
func New(appName string) *Notifier {
	return &Notifier{
		appName: appName,
		sem:     make(chan struct{}, 3),
	}
}

// ParseUrgency resolves the D-Bus urgency byte (0: low, 1: normal, 2: critical)
// and notification expiration timeout in milliseconds.
func ParseUrgency(urgency string) (byte, int32) {
	switch strings.ToLower(strings.TrimSpace(urgency)) {
	case "low":
		return 0, 5000
	case "critical":
		return 2, 0 // Critical alerts persist until dismissed by user
	default:
		return 1, 8000
	}
}

// Send emits a desktop notification with default normal urgency.
func (n *Notifier) Send(summary, body string) error {
	return n.SendWithUrgency(summary, body, "normal")
}

// SendWithUrgency emits a desktop notification with the specified urgency ("low", "normal", "critical").
func (n *Notifier) SendWithUrgency(summary, body, urgency string) error {
	select {
	case n.sem <- struct{}{}:
		defer func() { <-n.sem }()
	default:
		return errors.New("notification dropped: max concurrent notifications in flight")
	}

	conn, err := dbus.SessionBus()
	if err != nil {
		return fmt.Errorf("failed to connect to session bus: %w", err)
	}

	obj := conn.Object("org.freedesktop.Notifications", "/org/freedesktop/Notifications")

	urgencyVal, expireTimeout := ParseUrgency(urgency)

	// D-Bus org.freedesktop.Notifications.Notify signature:
	// (susssasa{sv}i) -> (u)
	hints := map[string]dbus.Variant{
		"urgency": dbus.MakeVariant(urgencyVal),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	call := obj.CallWithContext(
		ctx,
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
