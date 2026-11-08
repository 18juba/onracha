package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORSAllowsFrontendAPIMethods(t *testing.T) {
	api := NewRouter(nil, nil, nil, nil, Config{AllowedOrigin: "http://localhost:5173"})
	for _, method := range []string{http.MethodPut, http.MethodPatch, http.MethodDelete} {
		path := "/api/v1/arenas/1"
		if method == http.MethodPatch {
			path += "/active"
		} else if method == http.MethodDelete {
			path = "/api/v1/games/1/participants/me"
		}
		request := httptest.NewRequest(http.MethodOptions, path, nil)
		request.Header.Set("Origin", "http://localhost:5173")
		request.Header.Set("Access-Control-Request-Method", method)
		response := httptest.NewRecorder()

		api.ServeHTTP(response, request)

		if response.Code != http.StatusNoContent {
			t.Errorf("preflight status for %s = %d, want %d", method, response.Code, http.StatusNoContent)
		}
		if allowedMethods := response.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(allowedMethods, method) {
			t.Errorf("Access-Control-Allow-Methods = %q, want it to include %s", allowedMethods, method)
		}
	}
}
