package notify

import (
	"testing"
)

func TestParseUrgency(t *testing.T) {
	tests := []struct {
		input       string
		wantUrgency byte
		wantTimeout int32
	}{
		{"low", 0, 5000},
		{"LOW", 0, 5000},
		{"normal", 1, 8000},
		{"", 1, 8000},
		{"unknown", 1, 8000},
		{"critical", 2, 0},
		{"CRITICAL", 2, 0},
	}

	for _, tt := range tests {
		u, to := ParseUrgency(tt.input)
		if u != tt.wantUrgency || to != tt.wantTimeout {
			t.Errorf("ParseUrgency(%q) = (%d, %d), want (%d, %d)",
				tt.input, u, to, tt.wantUrgency, tt.wantTimeout)
		}
	}
}

func TestNotifier_SemaphoreLimit(t *testing.T) {
	n := New("TestApp")

	// Saturate all 3 semaphore slots
	n.sem <- struct{}{}
	n.sem <- struct{}{}
	n.sem <- struct{}{}

	// 4th notification attempt must fail immediately due to capacity limit
	err := n.SendWithUrgency("Test", "Body", "normal")
	if err == nil {
		t.Fatalf("expected error when semaphore is full, got nil")
	}
	if err.Error() != "notification dropped: max concurrent notifications in flight" {
		t.Errorf("unexpected error message: %v", err)
	}

	// Release 1 slot
	<-n.sem

	// Now capacity is available again, should proceed to D-Bus attempt (which may fail or succeed depending on D-Bus daemon)
	// We verify that the semaphore slot can be acquired without getting the "notification dropped" error
	select {
	case n.sem <- struct{}{}:
		<-n.sem
	default:
		t.Errorf("expected semaphore to have free capacity after release")
	}
}
