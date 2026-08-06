package auth

import "testing"

func TestNewTokenReturnsRawTokenAndOnlyItsHash(t *testing.T) {
	raw, hash, err := NewToken()
	if err != nil {
		t.Fatal(err)
	}
	if raw == "" || hash == "" || raw == hash {
		t.Fatal("invalid token")
	}
	if HashToken(raw) != hash {
		t.Fatal("token hash mismatch")
	}
}
