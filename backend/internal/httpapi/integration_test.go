package httpapi

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"onracha/internal/auth"
	"onracha/internal/postgres"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPostgreSQLRegistrationLoginAndCurrentAccount(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is not set; PostgreSQL integration test is skipped")
	}

	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open PostgreSQL connection: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := database.PingContext(ctx); err != nil {
		t.Fatalf("connect to PostgreSQL: %v", err)
	}

	email := fmt.Sprintf("onracha-%d@example.test", time.Now().UnixNano())
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := database.ExecContext(cleanupContext, "DELETE FROM users WHERE email = $1", email); err != nil {
			t.Errorf("clean up integration test account: %v", err)
		}
	})

	service := auth.NewService(postgres.NewAuthRepository(database), auth.NewArgon2idHasher())
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	api := NewRouter(service, logger, Config{AllowedOrigin: "http://localhost:5173"})

	registerRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(`{"name":"Integration Player","email":"`+email+`","password":"uma senha segura para teste"}`))
	registerRequest.Header.Set("Content-Type", "application/json")
	registerResponse := httptest.NewRecorder()
	api.ServeHTTP(registerResponse, registerRequest)
	if registerResponse.Code != http.StatusCreated {
		t.Fatalf("registration status = %d, want %d; body: %s", registerResponse.Code, http.StatusCreated, registerResponse.Body.String())
	}
	var registeredUser auth.User
	if err := json.Unmarshal(registerResponse.Body.Bytes(), &registeredUser); err != nil {
		t.Fatalf("decode registration response: %v", err)
	}
	if registeredUser.ID == 0 || registeredUser.Email != email {
		t.Fatalf("registration response user = %+v, want persisted public user", registeredUser)
	}
	if bytes.Contains(registerResponse.Body.Bytes(), []byte("password_hash")) {
		t.Fatal("registration response must not expose password hash")
	}

	duplicateRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(`{"name":"Another Player","email":"`+email+`","password":"uma senha segura para teste"}`))
	duplicateRequest.Header.Set("Content-Type", "application/json")
	duplicateResponse := httptest.NewRecorder()
	api.ServeHTTP(duplicateResponse, duplicateRequest)
	if duplicateResponse.Code != http.StatusConflict {
		t.Fatalf("duplicate registration status = %d, want %d", duplicateResponse.Code, http.StatusConflict)
	}

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"`+email+`","password":"uma senha segura para teste"}`))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	api.ServeHTTP(loginResponse, loginRequest)
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("login status = %d, want %d; body: %s", loginResponse.Code, http.StatusOK, loginResponse.Body.String())
	}
	cookies := loginResponse.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != sessionCookieName || !cookies[0].HttpOnly {
		t.Fatalf("login cookies = %+v, want one HttpOnly session cookie", cookies)
	}

	meRequest := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meRequest.AddCookie(cookies[0])
	meResponse := httptest.NewRecorder()
	api.ServeHTTP(meResponse, meRequest)
	if meResponse.Code != http.StatusOK {
		t.Fatalf("current account status = %d, want %d; body: %s", meResponse.Code, http.StatusOK, meResponse.Body.String())
	}
	var currentUser auth.User
	if err := json.Unmarshal(meResponse.Body.Bytes(), &currentUser); err != nil {
		t.Fatalf("decode current account response: %v", err)
	}
	if currentUser.ID != registeredUser.ID {
		t.Errorf("current account ID = %d, want %d", currentUser.ID, registeredUser.ID)
	}
}
