package arena

import (
	"errors"
	"math"
	"net/url"
	"strings"
	"time"
)

var (
	ErrInvalidOwner    = errors.New("arena owner is invalid")
	ErrInvalidArenaID  = errors.New("arena ID is invalid")
	ErrNotFound        = errors.New("arena not found")
	ErrInvalidName     = errors.New("arena name is required")
	ErrInvalidType     = errors.New("arena type is required")
	ErrInvalidAddress  = errors.New("arena address is required")
	ErrInvalidCity     = errors.New("arena city is required")
	ErrInvalidState    = errors.New("arena state must be a two-letter code")
	ErrInvalidLocation = errors.New("arena latitude and longitude must be valid coordinates")
	ErrInvalidPhotoURL = errors.New("arena photo URL must be an absolute HTTP or HTTPS URL")
)

// Arena is a sports venue managed by one ONRACHA user.
type Arena struct {
	ID          int64     `json:"id"`
	OwnerUserID int64     `json:"owner_user_id"`
	Name        string    `json:"name"`
	ArenaType   string    `json:"arena_type"`
	Address     string    `json:"address"`
	City        string    `json:"city"`
	State       string    `json:"state"`
	Latitude    *float64  `json:"latitude"`
	Longitude   *float64  `json:"longitude"`
	PhotoURL    *string   `json:"photo_url"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
}

// Details contains the venue fields an owner provides when creating or updating an arena.
type Details struct {
	Name      string
	ArenaType string
	Address   string
	City      string
	State     string
	Latitude  *float64
	Longitude *float64
}

// NewArena validates and normalizes the owner and venue details.
func NewArena(ownerUserID int64, details Details) (Arena, error) {
	if ownerUserID <= 0 {
		return Arena{}, ErrInvalidOwner
	}

	name := strings.TrimSpace(details.Name)
	if name == "" {
		return Arena{}, ErrInvalidName
	}

	arenaType := strings.ToLower(strings.TrimSpace(details.ArenaType))
	if arenaType == "" {
		return Arena{}, ErrInvalidType
	}

	address := strings.TrimSpace(details.Address)
	if address == "" {
		return Arena{}, ErrInvalidAddress
	}

	city := strings.TrimSpace(details.City)
	if city == "" {
		return Arena{}, ErrInvalidCity
	}

	state := strings.ToUpper(strings.TrimSpace(details.State))
	if !isStateCode(state) {
		return Arena{}, ErrInvalidState
	}
	if !validCoordinates(details.Latitude, details.Longitude) {
		return Arena{}, ErrInvalidLocation
	}
	return Arena{
		OwnerUserID: ownerUserID,
		Name:        name,
		ArenaType:   arenaType,
		Address:     address,
		City:        city,
		State:       state,
		Latitude:    details.Latitude,
		Longitude:   details.Longitude,
		Active:      true,
	}, nil
}

func validCoordinates(latitude, longitude *float64) bool {
	if (latitude == nil) != (longitude == nil) {
		return false
	}
	if latitude == nil {
		return true
	}
	return !math.IsNaN(*latitude) && !math.IsInf(*latitude, 0) &&
		!math.IsNaN(*longitude) && !math.IsInf(*longitude, 0) &&
		*latitude >= -90 && *latitude <= 90 && *longitude >= -180 && *longitude <= 180
}

func normalizePhotoURL(photoURL *string) (*string, error) {
	if photoURL == nil {
		return nil, nil
	}

	value := strings.TrimSpace(*photoURL)
	if value == "" {
		return nil, nil
	}

	parsedURL, err := url.Parse(value)
	if err != nil || parsedURL.Hostname() == "" ||
		(!strings.EqualFold(parsedURL.Scheme, "http") && !strings.EqualFold(parsedURL.Scheme, "https")) {
		return nil, ErrInvalidPhotoURL
	}
	return &value, nil
}

func isStateCode(state string) bool {
	if len(state) != 2 {
		return false
	}
	for _, character := range state {
		if character < 'A' || character > 'Z' {
			return false
		}
	}
	return true
}
