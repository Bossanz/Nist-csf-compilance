package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"
	"compliance/api/internal/httpapi"
	"compliance/api/internal/store"
)

func main() {
	ctx:=context.Background(); dbURL:=os.Getenv("DATABASE_URL"); if dbURL=="" { log.Fatal("DATABASE_URL is required") }
	s,err:=store.New(ctx,dbURL); if err!=nil { log.Fatal(err) }; defer s.Close()
	port:=os.Getenv("PORT"); if port=="" { port="8080" }
	server:=&http.Server{Addr:":"+port,Handler:httpapi.New(s),ReadHeaderTimeout:5*time.Second,ReadTimeout:15*time.Second,WriteTimeout:15*time.Second,IdleTimeout:60*time.Second}
	log.Printf("api listening on :%s",port); log.Fatal(server.ListenAndServe())
}
