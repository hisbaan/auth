package main

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"auth/internal/admin"
	"auth/internal/auth"
	"auth/internal/emails"
	internalMiddleware "auth/internal/middleware"
	"auth/internal/migrations"
	"auth/internal/oidc"
	"auth/internal/roles"
	"auth/internal/users"
	"auth/internal/wellknown"

	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
)

type Config struct {
	DatabaseURL         string `env:"DATABASE_URL,required"`
	BaseURL             string `env:"BASE_URL,required"`
	Port                string `env:"PORT,required"`
	JWTSigningKeyPEM    string `env:"JWT_SIGNING_KEY_FILE,file,required"`
	JWTSigningKeyID     string `env:"JWT_SIGNING_KEY_ID,required"`
	IssuerUrl           string `env:"ISSUER_URL,required"`
	ResendAPIKey        string `env:"RESEND_API_KEY,required"`
	FromEmail           string `env:"FROM_EMAIL,required"`
	FrontendURL         string `env:"FRONTEND_URL,required"`
	ServiceName         string `env:"SERVICE_NAME,required"`
	SupportEmail        string `env:"SUPPORT_EMAIL,required"`
	CookieDomain        string `env:"COOKIE_DOMAIN,required"`
	CORSOrigins         string `env:"CORS_ALLOWED_ORIGINS"`
	BlockedEmailDomains string `env:"BLOCKED_EMAIL_DOMAINS"`
	RunMigrations       bool   `env:"RUN_MIGRATIONS" envDefault:"true"`
}

func corsAllowedOrigins(cfg Config) []string {
	return splitCSV(cfg.CORSOrigins)
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
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

//	@title						auth
//	@version					0.0.1
//	@description				api for auth.hisbaan.com
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.
//
// TODO: Add response examples to API endpoints
// TODO: Add request body examples to API endpoints
func main() {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Setup db connection
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Duration(5) * time.Minute)
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("failed connecting to postgres: %v", err)
	}

	if cfg.RunMigrations {
		log.Println("applying database migrations")
		if err := migrations.Apply(context.Background(), db); err != nil {
			log.Fatalf("failed applying database migrations: %v", err)
		}
		log.Println("database migrations applied")
	} else {
		log.Println("database migrations disabled")
	}

	// Parse Ed25519 private key from PEM content
	signingKey, err := parseEd25519PrivateKey(cfg.JWTSigningKeyPEM)
	if err != nil {
		log.Fatalf("failed to parse signing key: %v", err)
	}

	// Setup resend
	emailService, err := emails.NewEmailService(cfg.ResendAPIKey, cfg.FromEmail, cfg.FrontendURL, cfg.ServiceName, cfg.SupportEmail)
	if err != nil {
		log.Fatalf("Failed to setup resend: %v", err)
	}

	// Setup chi router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(internalMiddleware.ClientInfo())
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(internalMiddleware.CORS(corsAllowedOrigins(cfg)))

	authService := auth.NewAuthService(db, signingKey, cfg.JWTSigningKeyID, cfg.IssuerUrl, emailService, cfg.CookieDomain, splitCSV(cfg.BlockedEmailDomains))
	r.Mount("/auth", auth.Router(authService))

	oidcService := oidc.NewOIDCService(db, signingKey, cfg.JWTSigningKeyID, cfg.IssuerUrl, cfg.FrontendURL, emailService, cfg.CookieDomain)
	r.Mount("/", oidc.Router(oidcService, signingKey.Public().(ed25519.PublicKey), cfg.IssuerUrl))

	usersService := users.NewUsersService(db, signingKey, cfg.IssuerUrl, emailService, splitCSV(cfg.BlockedEmailDomains))
	r.Mount("/users", users.Router(usersService))

	rolesService := roles.NewRolesService(db)
	r.Mount("/roles", roles.Router(rolesService))

	adminService := admin.NewAdminService(db)
	r.Mount("/admin", admin.Router(adminService, signingKey.Public().(ed25519.PublicKey), cfg.IssuerUrl))

	wellknownService := wellknown.NewWellKnownService(
		cfg.BaseURL,
		signingKey.Public().(ed25519.PublicKey),
		cfg.JWTSigningKeyID,
	)
	r.Mount("/.well-known", wellknown.Router(wellknownService))

	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
