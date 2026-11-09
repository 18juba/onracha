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
	"net/url"
	"os"
	"testing"
	"time"

	"onracha/internal/arena"
	"onracha/internal/auth"
	"onracha/internal/photostorage"
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

	photoDirectory := t.TempDir()
	photoStorage, err := photostorage.NewLocalStorage(photoDirectory, "http://localhost:8080")
	if err != nil {
		t.Fatalf("create local photo storage: %v", err)
	}
	arenaService := arena.NewService(postgres.NewArenaRepository(database), photoStorage)
	api := NewRouter(authService, arenaService, nil, slog.New(slog.NewTextHandler(io.Discard, nil)), Config{PhotoDirectory: photoDirectory})
	requestBody := `{"name":"Arena Central","arena_type":"Society","address":"Rua do Esporte, 100","city":"São Paulo","state":"sp","latitude":-23.55052,"longitude":-46.633308}`

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
	if createdArena.ID == 0 || createdArena.OwnerUserID != owner.ID || createdArena.ArenaType != "society" || createdArena.State != "SP" || createdArena.Latitude == nil || *createdArena.Latitude != -23.55052 || createdArena.Longitude == nil || *createdArena.Longitude != -46.633308 || createdArena.PhotoURL != nil {
		t.Fatalf("created arena = %+v, want normalized arena owned by authenticated user %d", createdArena, owner.ID)
	}

	photoBody, photoContentType := photoMultipartBody(t, encodeHTTPTestPNG(t), "text/plain")
	unauthenticatedPhotoRequest := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/arenas/%d/photo", createdArena.ID), photoBody)
	unauthenticatedPhotoRequest.Header.Set("Content-Type", photoContentType)
	unauthenticatedPhotoResponse := httptest.NewRecorder()
	api.ServeHTTP(unauthenticatedPhotoResponse, unauthenticatedPhotoRequest)
	if unauthenticatedPhotoResponse.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated photo upload status = %d, want %d", unauthenticatedPhotoResponse.Code, http.StatusUnauthorized)
	}

	photoBody, photoContentType = photoMultipartBody(t, encodeHTTPTestPNG(t), "text/plain")
	photoRequest := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/arenas/%d/photo", createdArena.ID), photoBody)
	photoRequest.Header.Set("Content-Type", photoContentType)
	photoRequest.AddCookie(&http.Cookie{Name: sessionCookieName, Value: session.Token})
	photoResponse := httptest.NewRecorder()
	api.ServeHTTP(photoResponse, photoRequest)
	if photoResponse.Code != http.StatusOK {
		t.Fatalf("photo upload status = %d, want %d; body: %s", photoResponse.Code, http.StatusOK, photoResponse.Body.String())
	}
	var uploadedArena arena.Arena
	if err := json.Unmarshal(photoResponse.Body.Bytes(), &uploadedArena); err != nil {
		t.Fatalf("decode arena after photo upload: %v", err)
	}
	if uploadedArena.PhotoURL == nil || uploadedArena.ID != createdArena.ID {
		t.Fatalf("arena after photo upload = %+v, want arena with stored photo reference", uploadedArena)
	}
	photoURL, err := url.Parse(*uploadedArena.PhotoURL)
	if err != nil {
		t.Fatalf("parse uploaded photo URL: %v", err)
	}
	photoReadResponse := httptest.NewRecorder()
	api.ServeHTTP(photoReadResponse, httptest.NewRequest(http.MethodGet, photoURL.Path, nil))
	if photoReadResponse.Code != http.StatusOK || photoReadResponse.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("read uploaded photo status/content type = %d/%q, want 200/image/png", photoReadResponse.Code, photoReadResponse.Header().Get("Content-Type"))
	}

	var persistedOwnerID int64
	var persistedLatitude, persistedLongitude sql.NullFloat64
	var persistedPhotoURL sql.NullString
	if err := database.QueryRowContext(ctx, "SELECT owner_user_id, latitude, longitude, photo_url FROM arenas WHERE id = $1", createdArena.ID).Scan(&persistedOwnerID, &persistedLatitude, &persistedLongitude, &persistedPhotoURL); err != nil {
		t.Fatalf("query persisted arena owner: %v", err)
	}
	if persistedOwnerID != owner.ID || !persistedLatitude.Valid || persistedLatitude.Float64 != -23.55052 || !persistedLongitude.Valid || persistedLongitude.Float64 != -46.633308 || !persistedPhotoURL.Valid || persistedPhotoURL.String != *uploadedArena.PhotoURL {
		t.Errorf("persisted arena fields = (%d, %v, %v, %v), want authenticated owner, coordinates and uploaded photo reference", persistedOwnerID, persistedLatitude, persistedLongitude, persistedPhotoURL)
	}

	invalidLocationRequest := httptest.NewRequest(http.MethodPost, "/api/v1/arenas", bytes.NewBufferString(`{"name":"Invalid Arena","arena_type":"society","address":"Rua Um","city":"São Paulo","state":"SP","latitude":-23.5}`))
	invalidLocationRequest.Header.Set("Content-Type", "application/json")
	invalidLocationRequest.AddCookie(&http.Cookie{Name: sessionCookieName, Value: session.Token})
	invalidLocationResponse := httptest.NewRecorder()
	api.ServeHTTP(invalidLocationResponse, invalidLocationRequest)
	if invalidLocationResponse.Code != http.StatusBadRequest {
		t.Errorf("create arena with incomplete coordinates status = %d, want %d; body: %s", invalidLocationResponse.Code, http.StatusBadRequest, invalidLocationResponse.Body.String())
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
