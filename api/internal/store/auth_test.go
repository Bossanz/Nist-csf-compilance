package store

import (
	"database/sql"
	"errors"
	"testing"
)

type nullablePasswordUserScanner struct{}

func (nullablePasswordUserScanner) Scan(dest ...any) error {
	if _, ok := dest[7].(*sql.NullString); !ok {
		return errors.New("password hash must be scanned as nullable text")
	}
	*dest[0].(*string) = "user-1"
	*dest[1].(**string) = nil
	*dest[2].(*string) = "Counselor"
	*dest[3].(*string) = "counselor@example.com"
	*dest[4].(*string) = "counselor"
	*dest[5].(*string) = "counselor"
	*dest[6].(*string) = "active"
	*dest[7].(*sql.NullString) = sql.NullString{}
	return nil
}

func TestScanUserAcceptsNullPasswordHash(t *testing.T) {
	user, err := scanUser(nullablePasswordUserScanner{})
	if err != nil {
		t.Fatalf("scanUser returned an error: %v", err)
	}
	if user.PasswordHash != "" {
		t.Fatalf("expected an empty password hash, got %q", user.PasswordHash)
	}
}
