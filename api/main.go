package main

import (
	"crypto/ed25519"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"time"

	"auth/internal/auth"

	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
)

type Config struct {
	DatabaseUrl      string `env:"DATABASE_URL,required"`
	Port             string `env:"PORT,required"`
	JWTAccessKeyPEM  string `env:"JWT_ACCESS_KEY_FILE,file,required"`
	JWTRefreshKeyPEM string `env:"JWT_REFRESH_KEY_FILE,file,required"`
	IssuerUrl        string `env:"ISSUER_URL,required"`
}

func parseEd25519PrivateKey(pemContent string) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemContent))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	edKey, ok := privateKey.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key is not an Ed25519 private key")
	}

	return edKey, nil
}

func main() {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Setup db connection
	db, err := sql.Open("postgres", cfg.DatabaseUrl)
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	defer db.Close()

	// Parse Ed25519 private keys from PEM content
	accessKey, err := parseEd25519PrivateKey(cfg.JWTAccessKeyPEM)
	if err != nil {
		log.Fatalf("failed to parse access key: %v", err)
	}
	refreshKey, err := parseEd25519PrivateKey(cfg.JWTRefreshKeyPEM)
	if err != nil {
		log.Fatalf("failed to parse refresh key: %v", err)
	}

	// Setup chi router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	authService, err := auth.NewAuthService(db, accessKey, refreshKey, cfg.IssuerUrl)
	if err != nil {
		log.Fatalf("failed to create auth service: %v", err)
	}
	r.Mount("/auth", auth.Router(authService))

	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
