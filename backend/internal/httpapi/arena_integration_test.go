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

	"onracha/internal/arena"
	"onracha/internal/auth"
	"onracha/internal/postgres"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestAuthenticatedUserCanCreateArenaForTheirAccount(t *testing.T) {
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

	email := fmt.Sprintf("arena-api-%d@example.test", time.Now().UnixNano())
	authService := auth.NewService(postgres.NewAuthRepository(database), auth.NewArgon2idHasher())
	owner, err := authService.Register(ctx, "Arena Owner", email, "Senha123")
	if err != nil {
		t.Fatalf("register test owner: %v", err)
	}
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		if _, err := database.ExecContext(cleanupContext, "DELETE FROM arenas WHERE owner_user_id = $1", owner.ID); err != nil {
			t.Errorf("clean up test arenas: %v", err)
		}
		if _, err := database.ExecContext(cleanupContext, "DELETE FROM users WHERE id = $1", owner.ID); err != nil {
			t.Errorf("clean up test owner: %v", err)
		}
	})

	_, session, err := authService.Login(ctx, email, "Senha123")
	if err != nil {
		t.Fatalf("login test owner: %v", err)
	}

	arenaService := arena.NewService(postgres.NewArenaRepository(database))
	api := NewRouter(authService, arenaService, slog.New(slog.NewTextHandler(io.Discard, nil)), Config{})
	requestBody := `{"name":"Arena Central","arena_type":"Society","address":"Rua do Esporte, 100","city":"São Paulo","state":"sp"}`

	unauthenticatedRequest := httptest.NewRequest(http.MethodPost, "/api/v1/arenas", bytes.NewBufferString(requestBody))
	unauthenticatedRequest.Header.Set("Content-Type", "application/json")
	unauthenticatedResponse := httptest.NewRecorder()
	api.ServeHTTP(unauthenticatedResponse, unauthenticatedRequest)
	if unauthenticatedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated create status = %d, want %d", unauthenticatedResponse.Code, http.StatusUnauthorized)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/arenas", bytes.NewBufferString(requestBody))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: session.Token})
	response := httptest.NewRecorder()
	api.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("create arena status = %d, want %d; body: %s", response.Code, http.StatusCreated, response.Body.String())
	}

	var createdArena arena.Arena
	if err := json.Unmarshal(response.Body.Bytes(), &createdArena); err != nil {
		t.Fatalf("decode create arena response: %v", err)
	}
	if createdArena.ID == 0 || createdArena.OwnerUserID != owner.ID || createdArena.ArenaType != "society" || createdArena.State != "SP" {
		t.Fatalf("created arena = %+v, want normalized arena owned by authenticated user %d", createdArena, owner.ID)
	}

	var persistedOwnerID int64
	if err := database.QueryRowContext(ctx, "SELECT owner_user_id FROM arenas WHERE id = $1", createdArena.ID).Scan(&persistedOwnerID); err != nil {
		t.Fatalf("query persisted arena owner: %v", err)
	}
	if persistedOwnerID != owner.ID {
		t.Errorf("persisted arena owner ID = %d, want authenticated user ID %d", persistedOwnerID, owner.ID)
	}

	spoofedOwnerRequest := httptest.NewRequest(http.MethodPost, "/api/v1/arenas", bytes.NewBufferString(`{"name":"Another Arena","arena_type":"society","address":"Rua do Esporte, 200","city":"São Paulo","state":"SP","owner_user_id":999999}`))
	spoofedOwnerRequest.Header.Set("Content-Type", "application/json")
	spoofedOwnerRequest.AddCookie(&http.Cookie{Name: sessionCookieName, Value: session.Token})
	spoofedOwnerResponse := httptest.NewRecorder()
	api.ServeHTTP(spoofedOwnerResponse, spoofedOwnerRequest)
	if spoofedOwnerResponse.Code != http.StatusBadRequest {
		t.Errorf("request with owner_user_id status = %d, want %d", spoofedOwnerResponse.Code, http.StatusBadRequest)
	}
}
