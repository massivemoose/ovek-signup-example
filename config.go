package main

import (
	"os"
	"strings"
	"time"
)

type config struct {
	Port             string
	PocketBaseURL    string
	SuperuserEmail   string
	SuperuserPass    string
	SuperuserToken   string
	RequestTimeout   time.Duration
	CollectionEnsure bool
}

func loadConfig() config {
	return config{
		Port:             env("PORT", "8080"),
		PocketBaseURL:    strings.TrimRight(env("POCKETBASE_URL", "http://127.0.0.1:8090"), "/"),
		SuperuserEmail:   os.Getenv("PB_SUPERUSER_EMAIL"),
		SuperuserPass:    os.Getenv("PB_SUPERUSER_PASSWORD"),
		SuperuserToken:   os.Getenv("PB_SUPERUSER_TOKEN"),
		RequestTimeout:   10 * time.Second,
		CollectionEnsure: true,
	}
}

func env(name string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}
