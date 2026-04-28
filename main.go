package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

func main() {
	cfg := loadConfig()
	pb := newPocketBaseClient(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if cfg.CollectionEnsure {
		if err := pb.ensureSignupsCollection(ctx); err != nil {
			log.Fatalf("ensure PocketBase collection: %v", err)
		}
	}

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           routes(pb),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("listening on :%s", cfg.Port)
	log.Fatal(server.ListenAndServe())
}
