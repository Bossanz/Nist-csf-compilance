package auth

import "testing"

func TestPasswordHashDoesNotStorePlaintext(t *testing.T) {
	password := "correct horse battery staple"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	if hash == password || !CheckPassword(hash, password) {
		t.Fatal("password hash verification failed")
	}
	if CheckPassword(hash, "wrong") {
		t.Fatal("wrong password matched")
	}
}
