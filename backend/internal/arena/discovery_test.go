package arena

import (
	"context"
	"errors"
	"math"
	"testing"
)

func TestServiceDiscoverTrimsCityAndReturnsPublicArenas(t *testing.T) {
	want := []PublicArena{{ID: 42, Name: "Arena Central", City: "São Paulo", State: "SP"}}
	repository := &fakeRepository{discoveredArenas: want}
	service := NewService(repository)

	got, err := service.Discover(context.Background(), DiscoveryFilters{City: "  São Paulo  "})
	if err != nil {
		t.Fatalf("Discover() error = %v, want nil", err)
	}
	if repository.discoverCalls != 1 || repository.discoverFilters.City != "São Paulo" {
		t.Errorf("repository discovery calls = %d, filters = %+v; want one call with trimmed city", repository.discoverCalls, repository.discoverFilters)
	}
	if len(got) != 1 || got[0] != want[0] {
		t.Errorf("Discover() = %+v, want %+v", got, want)
	}
}

func TestServiceDiscoverDefaultsLocationRadius(t *testing.T) {
	latitude, longitude := -23.55, -46.63
	repository := &fakeRepository{}
	service := NewService(repository)

	_, err := service.Discover(context.Background(), DiscoveryFilters{
		Latitude:  &latitude,
		Longitude: &longitude,
	})
	if err != nil {
		t.Fatalf("Discover() error = %v, want nil", err)
	}
	if repository.discoverFilters.RadiusKM == nil || *repository.discoverFilters.RadiusKM != DefaultDiscoveryRadiusKM {
		t.Errorf("repository radius = %v, want default %.0f km", repository.discoverFilters.RadiusKM, DefaultDiscoveryRadiusKM)
	}
}

func TestServiceDiscoverReturnsEmptySliceWhenNoArenasMatch(t *testing.T) {
	service := NewService(&fakeRepository{})

	got, err := service.Discover(context.Background(), DiscoveryFilters{City: "São Paulo"})
	if err != nil {
		t.Fatalf("Discover() error = %v, want nil", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("Discover() = %+v, want a non-nil empty slice", got)
	}
}

func TestServiceDiscoverRejectsInvalidFilters(t *testing.T) {
	latitude, longitude := -23.55, -46.63
	zeroRadius, tooLargeRadius := 0.0, MaxDiscoveryRadiusKM+1
	invalidLatitude, infiniteLongitude := 91.0, math.Inf(1)
	tests := []struct {
		name    string
		filters DiscoveryFilters
	}{
		{name: "no search criteria"},
		{name: "latitude without longitude", filters: DiscoveryFilters{Latitude: &latitude}},
		{name: "longitude without latitude", filters: DiscoveryFilters{Longitude: &longitude}},
		{name: "coordinates outside geographic bounds", filters: DiscoveryFilters{Latitude: &invalidLatitude, Longitude: &longitude}},
		{name: "non-finite coordinates", filters: DiscoveryFilters{Latitude: &latitude, Longitude: &infiniteLongitude}},
		{name: "radius without location", filters: DiscoveryFilters{City: "São Paulo", RadiusKM: &tooLargeRadius}},
		{name: "zero radius", filters: DiscoveryFilters{Latitude: &latitude, Longitude: &longitude, RadiusKM: &zeroRadius}},
		{name: "radius above maximum", filters: DiscoveryFilters{Latitude: &latitude, Longitude: &longitude, RadiusKM: &tooLargeRadius}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeRepository{}
			service := NewService(repository)
			if _, err := service.Discover(context.Background(), test.filters); !errors.Is(err, ErrInvalidDiscoveryFilters) {
				t.Errorf("Discover() error = %v, want %v", err, ErrInvalidDiscoveryFilters)
			}
			if repository.discoverCalls != 0 {
				t.Errorf("repository discovery calls = %d, want 0 for invalid filters", repository.discoverCalls)
			}
		})
	}
}

func TestServiceDiscoverReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	service := NewService(&fakeRepository{discoverError: wantErr})

	_, err := service.Discover(context.Background(), DiscoveryFilters{City: "São Paulo"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Discover() error = %v, want wrapped repository error %v", err, wantErr)
	}
}
