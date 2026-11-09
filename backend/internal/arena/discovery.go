package arena

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
)

const (
	DefaultDiscoveryRadiusKM = 25.0
	MaxDiscoveryRadiusKM     = 100.0
)

var ErrInvalidDiscoveryFilters = errors.New("arena discovery requires a city or valid location filters")

// DiscoveryFilters narrows public arena discovery by city and/or distance.
type DiscoveryFilters struct {
	City      string
	Latitude  *float64
	Longitude *float64
	RadiusKM  *float64
}

// PublicArena contains only the arena details needed for discovery.
type PublicArena struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	ArenaType string   `json:"arena_type"`
	Address   string   `json:"address"`
	City      string   `json:"city"`
	State     string   `json:"state"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
	PhotoURL  *string  `json:"photo_url"`
}

func normalizeDiscoveryFilters(filters DiscoveryFilters) (DiscoveryFilters, error) {
	filters.City = strings.TrimSpace(filters.City)
	if (filters.Latitude == nil) != (filters.Longitude == nil) {
		return DiscoveryFilters{}, ErrInvalidDiscoveryFilters
	}
	if filters.Latitude == nil {
		if filters.City == "" || filters.RadiusKM != nil {
			return DiscoveryFilters{}, ErrInvalidDiscoveryFilters
		}
		return filters, nil
	}

	latitude, longitude := *filters.Latitude, *filters.Longitude
	if math.IsNaN(latitude) || math.IsInf(latitude, 0) || latitude < -90 || latitude > 90 ||
		math.IsNaN(longitude) || math.IsInf(longitude, 0) || longitude < -180 || longitude > 180 {
		return DiscoveryFilters{}, ErrInvalidDiscoveryFilters
	}
	if filters.RadiusKM == nil {
		radius := DefaultDiscoveryRadiusKM
		filters.RadiusKM = &radius
	} else if math.IsNaN(*filters.RadiusKM) || math.IsInf(*filters.RadiusKM, 0) ||
		*filters.RadiusKM <= 0 || *filters.RadiusKM > MaxDiscoveryRadiusKM {
		return DiscoveryFilters{}, ErrInvalidDiscoveryFilters
	}
	return filters, nil
}

// Discover returns active public arenas matching the supplied search filters.
func (s *Service) Discover(ctx context.Context, filters DiscoveryFilters) ([]PublicArena, error) {
	normalizedFilters, err := normalizeDiscoveryFilters(filters)
	if err != nil {
		return nil, err
	}

	arenas, err := s.repository.Discover(ctx, normalizedFilters)
	if err != nil {
		return nil, fmt.Errorf("discover arenas: %w", err)
	}
	if arenas == nil {
		arenas = []PublicArena{}
	}
	return arenas, nil
}
