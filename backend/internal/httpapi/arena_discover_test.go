package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"onracha/internal/arena"
)

func TestArenaDiscoveryReturnsPublicArenasWithoutGames(t *testing.T) {
	api, repository := newPhotoUploadTestAPI(t)
	repository.discoveredArenas = []arena.PublicArena{{
		ID: 42, Name: "Arena Central", ArenaType: "society", Address: "Rua Um, 10",
		City: "São Paulo", State: "SP",
	}}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/arenas/discover?city=%20S%C3%A3o%20Paulo%20", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid-session"})
	response := httptest.NewRecorder()

	api.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("discovery status = %d, want %d; body: %s", response.Code, http.StatusOK, response.Body.String())
	}
	var discovered []arena.PublicArena
	if err := json.Unmarshal(response.Body.Bytes(), &discovered); err != nil {
		t.Fatalf("decode discovered arenas: %v", err)
	}
	if len(discovered) != 1 || discovered[0].ID != 42 || discovered[0].Name != "Arena Central" {
		t.Errorf("discovered arenas = %+v, want the matching public arena", discovered)
	}
	if repository.discoveryCalls != 1 || repository.discoveryFilters.City != "São Paulo" {
		t.Errorf("repository discovery calls = %d, city filter = %q; want one call with trimmed city", repository.discoveryCalls, repository.discoveryFilters.City)
	}
	if strings.Contains(response.Body.String(), "owner_user_id") {
		t.Error("public arena discovery response must not expose the owner user ID")
	}
}

func TestArenaDiscoveryPassesLocationAndRadiusFilters(t *testing.T) {
	api, repository := newPhotoUploadTestAPI(t)
	request := httptest.NewRequest(http.MethodGet,
		"/api/v1/arenas/discover?latitude=-23.55&longitude=-46.63&radius_km=12.5", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid-session"})
	response := httptest.NewRecorder()

	api.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("discovery status = %d, want %d; body: %s", response.Code, http.StatusOK, response.Body.String())
	}
	filters := repository.discoveryFilters
	if filters.Latitude == nil || *filters.Latitude != -23.55 || filters.Longitude == nil || *filters.Longitude != -46.63 ||
		filters.RadiusKM == nil || *filters.RadiusKM != 12.5 {
		t.Errorf("repository discovery filters = %+v, want requested coordinates and 12.5 km radius", filters)
	}
}

func TestArenaDiscoveryRejectsMissingOrInvalidFilters(t *testing.T) {
	tests := []struct {
		name  string
		query string
	}{
		{name: "no criteria"},
		{name: "latitude without longitude", query: "?latitude=-23.55"},
		{name: "invalid coordinate", query: "?latitude=nearby&longitude=-46.63"},
		{name: "radius without location", query: "?city=S%C3%A3o%20Paulo&radius_km=10"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			api, repository := newPhotoUploadTestAPI(t)
			request := httptest.NewRequest(http.MethodGet, "/api/v1/arenas/discover"+test.query, nil)
			request.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "valid-session"})
			response := httptest.NewRecorder()

			api.ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Errorf("discovery status = %d, want %d; body: %s", response.Code, http.StatusBadRequest, response.Body.String())
			}
			if repository.discoveryCalls != 0 {
				t.Errorf("repository discovery calls = %d, want 0 for invalid filters", repository.discoveryCalls)
			}
		})
	}
}

func TestArenaDiscoveryRequiresAuthentication(t *testing.T) {
	api, repository := newPhotoUploadTestAPI(t)
	request := httptest.NewRequest(http.MethodGet, "/api/v1/arenas/discover?city=S%C3%A3o%20Paulo", nil)
	response := httptest.NewRecorder()

	api.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Errorf("unauthenticated discovery status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if repository.discoveryCalls != 0 {
		t.Errorf("repository discovery calls = %d, want 0 for unauthenticated request", repository.discoveryCalls)
	}
}
