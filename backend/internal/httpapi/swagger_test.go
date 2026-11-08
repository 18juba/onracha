package httpapi

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSwaggerUIIsServedAtSwaggerRoute(t *testing.T) {
	router := newSwaggerTestRouter(t)
	request := httptest.NewRequest(http.MethodGet, "/swagger", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if contentType := response.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", contentType)
	}
	for _, expected := range []string{"SwaggerUIBundle", "/openapi.yaml"} {
		if !strings.Contains(response.Body.String(), expected) {
			t.Errorf("Swagger page does not contain %q", expected)
		}
	}
}

func TestOpenAPISpecIsServedAtOpenAPIRoute(t *testing.T) {
	router := newSwaggerTestRouter(t)
	request := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if contentType := response.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "application/yaml") {
		t.Errorf("Content-Type = %q, want application/yaml", contentType)
	}
	if got, want := response.Body.String(), "openapi: 3.1.0\n"; got != want {
		t.Errorf("body = %q, want %q", got, want)
	}
}

func TestOpenAPIRouteReturnsInternalErrorWhenSpecIsMissing(t *testing.T) {
	router := NewRouter(nil, nil, nil, slog.New(slog.NewTextHandler(io.Discard, nil)), Config{
		OpenAPISpecPath: filepath.Join(t.TempDir(), "missing-openapi.yaml"),
	})
	request := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
}

func newSwaggerTestRouter(t *testing.T) http.Handler {
	t.Helper()

	specPath := filepath.Join(t.TempDir(), "openapi.yaml")
	if err := os.WriteFile(specPath, []byte("openapi: 3.1.0\n"), 0o600); err != nil {
		t.Fatalf("write test OpenAPI document: %v", err)
	}

	return NewRouter(nil, nil, nil, slog.New(slog.NewTextHandler(io.Discard, nil)), Config{OpenAPISpecPath: specPath})
}
