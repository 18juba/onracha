package httpapi

import (
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

func TestAuthenticatedUserCanAccessOnlyTheirArenas(t *testing.T) {
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

	authService := auth.NewService(postgres.NewAuthRepository(database), auth.NewArgon2idHasher())
	owners := make([]auth.User, 0, 2)
	t.Cleanup(func() {
		cleanupContext, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		for _, owner := range owners {
			if _, err := database.ExecContext(cleanupContext, "DELETE FROM arenas WHERE owner_user_id = $1", owner.ID); err != nil {
				t.Errorf("clean up arenas for user %d: %v", owner.ID, err)
			}
			if _, err := database.ExecContext(cleanupContext, "DELETE FROM users WHERE id = $1", owner.ID); err != nil {
				t.Errorf("clean up user %d: %v", owner.ID, err)
			}
		}
	})

	firstOwner, err := authService.Register(ctx, "First Owner", fmt.Sprintf("arena-list-first-%d@example.test", time.Now().UnixNano()), "Senha123")
	if err != nil {
		t.Fatalf("register first owner: %v", err)
	}
	owners = append(owners, firstOwner)
	secondOwner, err := authService.Register(ctx, "Second Owner", fmt.Sprintf("arena-list-second-%d@example.test", time.Now().UnixNano()), "Senha123")
	if err != nil {
		t.Fatalf("register second owner: %v", err)
	}
	owners = append(owners, secondOwner)

	_, session, err := authService.Login(ctx, firstOwner.Email, "Senha123")
	if err != nil {
		t.Fatalf("login first owner: %v", err)
	}

	arenaService := arena.NewService(postgres.NewArenaRepository(database))
	firstArena, err := arenaService.Create(ctx, firstOwner.ID, arena.Details{
		Name: "First Owner Arena", ArenaType: "society", Address: "Rua Um, 10", City: "São Paulo", State: "SP",
	})
	if err != nil {
		t.Fatalf("create first owner's arena: %v", err)
	}
	if _, err := arenaService.Create(ctx, secondOwner.ID, arena.Details{
		Name: "Second Owner Arena", ArenaType: "futsal", Address: "Rua Dois, 20", City: "São Paulo", State: "SP",
	}); err != nil {
		t.Fatalf("create second owner's arena: %v", err)
	}

	api := NewRouter(authService, arenaService, slog.New(slog.NewTextHandler(io.Discard, nil)), Config{})
	unauthenticatedRequest := httptest.NewRequest(http.MethodGet, "/api/v1/arenas", nil)
	unauthenticatedResponse := httptest.NewRecorder()
	api.ServeHTTP(unauthenticatedResponse, unauthenticatedRequest)
	if unauthenticatedResponse.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated list status = %d, want %d", unauthenticatedResponse.Code, http.StatusUnauthorized)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/arenas", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: session.Token})
	response := httptest.NewRecorder()
	api.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("list arenas status = %d, want %d; body: %s", response.Code, http.StatusOK, response.Body.String())
	}

	var arenas []arena.Arena
	if err := json.Unmarshal(response.Body.Bytes(), &arenas); err != nil {
		t.Fatalf("decode list arenas response: %v", err)
	}
	if len(arenas) != 1 {
		t.Fatalf("listed arenas = %d, want 1; body: %s", len(arenas), response.Body.String())
	}
	if arenas[0].ID != firstArena.ID || arenas[0].OwnerUserID != firstOwner.ID {
		t.Errorf("listed arena = %+v, want first owner's arena %d", arenas[0], firstArena.ID)
	}

	detailPath := fmt.Sprintf("/api/v1/arenas/%d", firstArena.ID)
	unauthenticatedDetailRequest := httptest.NewRequest(http.MethodGet, detailPath, nil)
	unauthenticatedDetailResponse := httptest.NewRecorder()
	api.ServeHTTP(unauthenticatedDetailResponse, unauthenticatedDetailRequest)
	if unauthenticatedDetailResponse.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated detail status = %d, want %d", unauthenticatedDetailResponse.Code, http.StatusUnauthorized)
	}

	ownerDetailRequest := httptest.NewRequest(http.MethodGet, detailPath, nil)
	ownerDetailRequest.AddCookie(&http.Cookie{Name: sessionCookieName, Value: session.Token})
	ownerDetailResponse := httptest.NewRecorder()
	api.ServeHTTP(ownerDetailResponse, ownerDetailRequest)
	if ownerDetailResponse.Code != http.StatusOK {
		t.Fatalf("owner detail status = %d, want %d; body: %s", ownerDetailResponse.Code, http.StatusOK, ownerDetailResponse.Body.String())
	}

	var ownerArena arena.Arena
	if err := json.Unmarshal(ownerDetailResponse.Body.Bytes(), &ownerArena); err != nil {
		t.Fatalf("decode owner detail response: %v", err)
	}
	if ownerArena.ID != firstArena.ID || ownerArena.OwnerUserID != firstOwner.ID || ownerArena.Name != firstArena.Name || ownerArena.ArenaType != firstArena.ArenaType || ownerArena.Address != firstArena.Address || ownerArena.City != firstArena.City || ownerArena.State != firstArena.State || !ownerArena.CreatedAt.Equal(firstArena.CreatedAt) {
		t.Errorf("owner detail = %+v, want full arena %+v", ownerArena, firstArena)
	}

	invalidIDRequest := httptest.NewRequest(http.MethodGet, "/api/v1/arenas/not-an-integer", nil)
	invalidIDRequest.AddCookie(&http.Cookie{Name: sessionCookieName, Value: session.Token})
	invalidIDResponse := httptest.NewRecorder()
	api.ServeHTTP(invalidIDResponse, invalidIDRequest)
	if invalidIDResponse.Code != http.StatusBadRequest {
		t.Fatalf("invalid arena ID status = %d, want %d", invalidIDResponse.Code, http.StatusBadRequest)
	}

	_, secondSession, err := authService.Login(ctx, secondOwner.Email, "Senha123")
	if err != nil {
		t.Fatalf("login second owner: %v", err)
	}
	otherOwnerRequest := httptest.NewRequest(http.MethodGet, detailPath, nil)
	otherOwnerRequest.AddCookie(&http.Cookie{Name: sessionCookieName, Value: secondSession.Token})
	otherOwnerResponse := httptest.NewRecorder()
	api.ServeHTTP(otherOwnerResponse, otherOwnerRequest)
	if otherOwnerResponse.Code != http.StatusNotFound {
		t.Fatalf("other owner's detail status = %d, want %d", otherOwnerResponse.Code, http.StatusNotFound)
	}
}
