package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"compliance/api/internal/auth"
	"compliance/api/internal/store"
)

func validateInput(databaseURL, email, password string) error {
	if strings.TrimSpace(databaseURL) == "" {
		return fmt.Errorf("database URL is required")
	}
	if strings.TrimSpace(email) == "" {
		return fmt.Errorf("email is required")
	}
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	return nil
}

func main() {
	databaseURL := flag.String("database-url", os.Getenv("DATABASE_URL"), "PostgreSQL connection URL")
	email := flag.String("email", "", "active user email")
	password := flag.String("password", "", "new local password")
	flag.Parse()

	if err := validateInput(*databaseURL, *email, *password); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	data, err := store.New(ctx, *databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer data.Close()

	user, err := data.FindActiveUserByEmail(ctx, auth.NormalizeEmail(*email))
	if err != nil {
		log.Fatalf("find active user: %v", err)
	}
	hash, err := auth.HashPassword(*password)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}
	if err := data.ChangePassword(ctx, user.ID, hash); err != nil {
		log.Fatalf("reset password: %v", err)
	}
	fmt.Printf("password reset for %s\n", user.Email)
}
