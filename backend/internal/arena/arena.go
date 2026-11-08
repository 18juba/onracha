package arena

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidOwner   = errors.New("arena owner is invalid")
	ErrInvalidArenaID = errors.New("arena ID is invalid")
	ErrNotFound       = errors.New("arena not found")
	ErrInvalidName    = errors.New("arena name is required")
	ErrInvalidType    = errors.New("arena type is required")
	ErrInvalidAddress = errors.New("arena address is required")
	ErrInvalidCity    = errors.New("arena city is required")
	ErrInvalidState   = errors.New("arena state must be a two-letter code")
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
	CreatedAt   time.Time `json:"created_at"`
}

// Details contains the fields an owner provides when registering an arena.
type Details struct {
	Name      string
	ArenaType string
	Address   string
	City      string
	State     string
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

	return Arena{
		OwnerUserID: ownerUserID,
		Name:        name,
		ArenaType:   arenaType,
		Address:     address,
		City:        city,
		State:       state,
	}, nil
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
