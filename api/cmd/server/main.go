package main

import (
	"compliance/api/internal/auth"
	"compliance/api/internal/httpapi"
	"compliance/api/internal/store"
	"context"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	s, err := store.New(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()
	authService := auth.NewService(s, time.Now)
	adminEmail, adminPassword := os.Getenv("BOOTSTRAP_ADMIN_EMAIL"), os.Getenv("BOOTSTRAP_ADMIN_PASSWORD")
	if adminEmail != "" && adminPassword != "" {
		if err := authService.Bootstrap(ctx, adminEmail, adminPassword); err != nil {
			log.Fatal(err)
		}
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{Addr: ":" + port, Handler: httpapi.New(s, authService, os.Getenv("APP_ENV") == "production", os.Getenv("APP_ORIGIN")), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("api listening on :%s", port)
	log.Fatal(server.ListenAndServe())
}
