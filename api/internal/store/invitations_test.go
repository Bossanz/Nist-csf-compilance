package store

import (
	"testing"
	"time"
)

func TestInvitationStatus(t *testing.T) {
	now := time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)
	expired := now.Add(-time.Minute)
	future := now.Add(time.Hour)
	cancelled := now.Add(-time.Hour)
	superseded := now.Add(-30 * time.Minute)
	accepted := now.Add(-15 * time.Minute)

	tests := []struct {
		name string
		inv  Invitation
		want string
	}{
		{name: "pending", inv: Invitation{ExpiresAt: future}, want: "pending"},
		{name: "expired", inv: Invitation{ExpiresAt: expired}, want: "expired"},
		{name: "cancelled takes precedence over expiry", inv: Invitation{ExpiresAt: expired, CancelledAt: &cancelled}, want: "cancelled"},
		{name: "superseded takes precedence over expiry", inv: Invitation{ExpiresAt: expired, SupersededAt: &superseded}, want: "superseded"},
		{name: "accepted is terminal", inv: Invitation{ExpiresAt: expired, AcceptedAt: &accepted, CancelledAt: &cancelled}, want: "accepted"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InvitationStatus(tt.inv, now); got != tt.want {
				t.Fatalf("InvitationStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}
