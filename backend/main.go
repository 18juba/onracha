package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"onracha/internal/arena"
	"onracha/internal/auth"
	"onracha/internal/game"
	"onracha/internal/httpapi"
	"onracha/internal/postgres"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("api stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}

	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open database connection: %w", err)
	}
	defer database.Close()

	startupContext, cancelStartup := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelStartup()
	if err := database.PingContext(startupContext); err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}

	secureCookies, err := envBool("SESSION_COOKIE_SECURE", false)
	if err != nil {
		return err
	}
	frontendOrigin := os.Getenv("FRONTEND_ORIGIN")
	openAPISpecPath := os.Getenv("OPENAPI_SPEC_PATH")
	if frontendOrigin == "" {
		frontendOrigin = "http://localhost:5173"
	}

	authService := auth.NewService(postgres.NewAuthRepository(database), auth.NewArgon2idHasher())
	arenaService := arena.NewService(postgres.NewArenaRepository(database))
	gameService := game.NewService(postgres.NewTimeSlotRepository(database), postgres.NewGameRepository(database))
	server := &http.Server{
		Addr:              ":8080",
		Handler:           httpapi.NewRouter(authService, arenaService, gameService, logger, httpapi.Config{AllowedOrigin: frontendOrigin, SecureCookies: secureCookies, OpenAPISpecPath: openAPISpecPath}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	shutdownContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-shutdownContext.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("graceful shutdown failed", "error", err)
		}
	}()

	logger.Info("api listening", "address", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP: %w", err)
	}
	return nil
}

func envBool(name string, fallback bool) (bool, error) {
	value, exists := os.LookupEnv(name)
	if !exists || value == "" {
		return fallback, nil
	}
	parsedValue, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("parse %s: %w", name, err)
	}
	return parsedValue, nil
}
