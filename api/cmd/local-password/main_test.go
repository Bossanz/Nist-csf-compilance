package main

import "testing"

func TestValidateInputRequiresLocalResetArguments(t *testing.T) {
	tests := []struct {
		name        string
		databaseURL string
		email       string
		password    string
	}{
		{name: "database url", email: "admin@example.com", password: "LocalAdmin!2026"},
		{name: "email", databaseURL: "postgres://localhost/compliance", password: "LocalAdmin!2026"},
		{name: "password", databaseURL: "postgres://localhost/compliance", email: "admin@example.com"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := validateInput(test.databaseURL, test.email, test.password); err == nil {
				t.Fatal("expected missing argument error")
			}
		})
	}
}
