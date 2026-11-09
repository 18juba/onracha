package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"testing"

	"onracha/internal/arena"
	"onracha/internal/auth"
	"onracha/internal/photostorage"
)

func TestArenaPhotoUploadValidatesContentAndServesStoredImage(t *testing.T) {
	api, repository := newPhotoUploadTestAPI(t)
	imageBytes := encodeHTTPTestPNG(t)
	body, contentType := photoMultipartBody(t, imageBytes, "text/plain")
	request := httptest.NewRequest(http.MethodPut, "/api/v1/arenas/42/photo", body)
	request.Header.Set("Content-Type", contentType)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid-session"})
	response := httptest.NewRecorder()
	api.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("upload status = %d, want %d; body: %s", response.Code, http.StatusOK, response.Body.String())
	}

	var uploaded arena.Arena
	if err := json.Unmarshal(response.Body.Bytes(), &uploaded); err != nil {
		t.Fatalf("decode uploaded arena: %v", err)
	}
	if uploaded.PhotoURL == nil || repository.photoURL == nil || *uploaded.PhotoURL != *repository.photoURL {
		t.Fatalf("uploaded photo URL = %v, repository photo URL = %v; want matching stored reference", uploaded.PhotoURL, repository.photoURL)
	}

	photoURL, err := url.Parse(*uploaded.PhotoURL)
	if err != nil {
		t.Fatalf("parse uploaded photo URL: %v", err)
	}
	photoRequest := httptest.NewRequest(http.MethodGet, photoURL.Path, nil)
	photoResponse := httptest.NewRecorder()
	api.ServeHTTP(photoResponse, photoRequest)
	if photoResponse.Code != http.StatusOK {
		t.Fatalf("read uploaded photo status = %d, want %d", photoResponse.Code, http.StatusOK)
	}
	if got := photoResponse.Header().Get("Content-Type"); got != "image/png" {
		t.Errorf("served image Content-Type = %q, want image/png", got)
	}
	if !bytes.Equal(photoResponse.Body.Bytes(), imageBytes) {
		t.Error("served photo bytes differ from the uploaded PNG")
	}
}

func TestArenaPhotoUploadRejectsInvalidAndOversizedContent(t *testing.T) {
	api, repository := newPhotoUploadTestAPI(t)
	tests := []struct {
		name            string
		data            []byte
		partContentType string
		wantStatus      int
	}{
		{name: "invalid bytes labeled as PNG", data: []byte("not an image"), partContentType: "image/png", wantStatus: http.StatusBadRequest},
		{name: "oversized file", data: bytes.Repeat([]byte("x"), arena.MaxPhotoSize+1), partContentType: "image/png", wantStatus: http.StatusRequestEntityTooLarge},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, contentType := photoMultipartBody(t, test.data, test.partContentType)
			request := httptest.NewRequest(http.MethodPut, "/api/v1/arenas/42/photo", body)
			request.Header.Set("Content-Type", contentType)
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid-session"})
			response := httptest.NewRecorder()
			api.ServeHTTP(response, request)
			if response.Code != test.wantStatus {
				t.Errorf("upload status = %d, want %d; body: %s", response.Code, test.wantStatus, response.Body.String())
			}
		})
	}
	if repository.photoUpdateCalls != 0 {
		t.Errorf("repository photo updates = %d, want 0 for rejected files", repository.photoUpdateCalls)
	}
}

func TestArenaPhotoUploadRequiresAuthentication(t *testing.T) {
	api, repository := newPhotoUploadTestAPI(t)
	body, contentType := photoMultipartBody(t, encodeHTTPTestPNG(t), "image/png")
	request := httptest.NewRequest(http.MethodPut, "/api/v1/arenas/42/photo", body)
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()
	api.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated upload status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if repository.photoUpdateCalls != 0 {
		t.Errorf("repository photo updates = %d, want 0", repository.photoUpdateCalls)
	}
}

func TestArenaPhotoUploadIsRestrictedToTheArenaOwner(t *testing.T) {
	api, repository := newPhotoUploadTestAPI(t)
	body, contentType := photoMultipartBody(t, encodeHTTPTestPNG(t), "image/png")
	request := httptest.NewRequest(http.MethodPut, "/api/v1/arenas/42/photo", body)
	request.Header.Set("Content-Type", contentType)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "other-session"})
	response := httptest.NewRecorder()
	api.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Errorf("non-owner upload status = %d, want %d", response.Code, http.StatusNotFound)
	}
	if repository.photoUpdateCalls != 0 {
		t.Errorf("repository photo updates = %d, want 0", repository.photoUpdateCalls)
	}
}

func newPhotoUploadTestAPI(t *testing.T) (http.Handler, *photoUploadRepository) {
	t.Helper()
	photoDirectory := t.TempDir()
	storage, err := photostorage.NewLocalStorage(photoDirectory, "http://api.example.test")
	if err != nil {
		t.Fatalf("create local photo storage: %v", err)
	}
	repository := &photoUploadRepository{value: arena.Arena{ID: 42, OwnerUserID: 17, Name: "Arena Test"}}
	arenaService := arena.NewService(repository, storage)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	api := NewRouter(&photoUploadAuthService{}, arenaService, nil, logger, Config{PhotoDirectory: photoDirectory})
	return api, repository
}

func photoMultipartBody(t *testing.T, data []byte, partContentType string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Disposition", `form-data; name="file"; filename="arena.png"`)
	partHeader.Set("Content-Type", partContentType)
	part, err := writer.CreatePart(partHeader)
	if err != nil {
		t.Fatalf("create multipart file part: %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("write multipart file part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	return body, writer.FormDataContentType()
}

func encodeHTTPTestPNG(t *testing.T) []byte {
	t.Helper()
	var output bytes.Buffer
	if err := png.Encode(&output, image.NewRGBA(image.Rect(0, 0, 2, 2))); err != nil {
		t.Fatalf("encode test PNG: %v", err)
	}
	return output.Bytes()
}

type photoUploadAuthService struct{}

func (photoUploadAuthService) Register(context.Context, string, string, string) (auth.User, error) {
	return auth.User{}, errors.New("not used in photo upload test")
}

func (photoUploadAuthService) Login(context.Context, string, string) (auth.User, auth.Session, error) {
	return auth.User{}, auth.Session{}, errors.New("not used in photo upload test")
}

func (photoUploadAuthService) Authenticate(_ context.Context, token string) (auth.User, error) {
	switch token {
	case "valid-session":
		return auth.User{ID: 17}, nil
	case "other-session":
		return auth.User{ID: 18}, nil
	default:
		return auth.User{}, auth.ErrInvalidSession
	}
}

type photoUploadRepository struct {
	value            arena.Arena
	photoURL         *string
	photoUpdateCalls int
	discoveredArenas []arena.PublicArena
	discoveryFilters arena.DiscoveryFilters
	discoveryCalls   int
	discoveryError   error
}

func (r *photoUploadRepository) Create(_ context.Context, value arena.Arena) (arena.Arena, error) {
	return value, nil
}

func (r *photoUploadRepository) ListByOwner(context.Context, int64) ([]arena.Arena, error) {
	return []arena.Arena{}, nil
}

func (r *photoUploadRepository) Discover(_ context.Context, filters arena.DiscoveryFilters) ([]arena.PublicArena, error) {
	r.discoveryCalls++
	r.discoveryFilters = filters
	if r.discoveryError != nil {
		return nil, r.discoveryError
	}
	return r.discoveredArenas, nil
}

func (r *photoUploadRepository) GetByIDAndOwner(_ context.Context, arenaID, ownerUserID int64) (arena.Arena, error) {
	if arenaID != r.value.ID || ownerUserID != r.value.OwnerUserID {
		return arena.Arena{}, arena.ErrNotFound
	}
	return r.value, nil
}

func (r *photoUploadRepository) Update(_ context.Context, value arena.Arena) (arena.Arena, error) {
	return value, nil
}

func (r *photoUploadRepository) ToggleActive(context.Context, int64, int64) (arena.Arena, error) {
	return r.value, nil
}

func (r *photoUploadRepository) UpdatePhotoURL(_ context.Context, arenaID, ownerUserID int64, photoURL string) (arena.Arena, error) {
	if arenaID != r.value.ID || ownerUserID != r.value.OwnerUserID {
		return arena.Arena{}, arena.ErrNotFound
	}
	r.photoUpdateCalls++
	r.photoURL = &photoURL
	r.value.PhotoURL = &photoURL
	return r.value, nil
}
