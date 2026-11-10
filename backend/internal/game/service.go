package game

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Service validates scheduling requests and delegates persistence operations.
type Service struct {
	timeSlots TimeSlotRepository
	games     GameRepository
}

// NewService creates a game service with its persistence dependencies.
func NewService(timeSlots TimeSlotRepository, games GameRepository) *Service {
	return &Service{timeSlots: timeSlots, games: games}
}

// CreateTimeSlot registers a future arena interval on behalf of its owner.
func (s *Service) CreateTimeSlot(ctx context.Context, arenaID, ownerUserID int64, startsAt, endsAt time.Time) (TimeSlot, error) {
	if ownerUserID <= 0 {
		return TimeSlot{}, ErrInvalidUserID
	}

	value, err := NewTimeSlot(arenaID, startsAt, endsAt)
	if err != nil {
		return TimeSlot{}, err
	}

	created, err := s.timeSlots.Create(ctx, ownerUserID, value)
	if err != nil {
		return TimeSlot{}, fmt.Errorf("create arena time slot: %w", err)
	}
	return created, nil
}

// ListAvailableTimeSlots returns unbooked future intervals for one arena.
func (s *Service) ListAvailableTimeSlots(ctx context.Context, arenaID int64) ([]TimeSlot, error) {
	if arenaID <= 0 {
		return nil, ErrInvalidArenaID
	}

	values, err := s.timeSlots.ListAvailableByArena(ctx, arenaID)
	if err != nil {
		return nil, fmt.Errorf("list available arena time slots: %w", err)
	}
	if values == nil {
		values = []TimeSlot{}
	}
	return values, nil
}

// Schedule creates a game and automatically registers its organizer as a participant.
func (s *Service) Schedule(ctx context.Context, timeSlotID, organizerUserID int64) (Game, error) {
	if timeSlotID <= 0 {
		return Game{}, ErrInvalidTimeSlotID
	}
	if organizerUserID <= 0 {
		return Game{}, ErrInvalidUserID
	}

	created, err := s.games.Schedule(ctx, timeSlotID, organizerUserID)
	if err != nil {
		return Game{}, fmt.Errorf("schedule game: %w", err)
	}
	return created, nil
}

// ListByArena returns scheduled games and their participants for one arena.
func (s *Service) ListByArena(ctx context.Context, arenaID int64) ([]Game, error) {
	if arenaID <= 0 {
		return nil, ErrInvalidArenaID
	}

	values, err := s.games.ListByArena(ctx, arenaID)
	if err != nil {
		return nil, fmt.Errorf("list games by arena: %w", err)
	}
	if values == nil {
		values = []Game{}
	}
	return values, nil
}

// ListByParticipant returns future games in which the user is enrolled.
func (s *Service) ListByParticipant(ctx context.Context, userID int64) ([]Game, error) {
	if userID <= 0 {
		return nil, ErrInvalidUserID
	}

	values, err := s.games.ListByParticipant(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list games by participant: %w", err)
	}
	if values == nil {
		values = []Game{}
	}
	return values, nil
}

// GetByID returns one game with its participants.
func (s *Service) GetByID(ctx context.Context, gameID int64) (Game, error) {
	if gameID <= 0 {
		return Game{}, ErrInvalidGameID
	}

	value, err := s.games.GetByID(ctx, gameID)
	if err != nil {
		return Game{}, fmt.Errorf("get game by ID: %w", err)
	}
	return value, nil
}

// GetLobbyByID returns one game with arena details and displayable participant names.
func (s *Service) GetLobbyByID(ctx context.Context, gameID int64) (GameLobby, error) {
	if gameID <= 0 {
		return GameLobby{}, ErrInvalidGameID
	}

	value, err := s.games.GetLobbyByID(ctx, gameID)
	if err != nil {
		return GameLobby{}, fmt.Errorf("get game lobby by ID: %w", err)
	}
	return value, nil
}

// Join adds a user to a game unless the user already has an overlapping game.
func (s *Service) Join(ctx context.Context, gameID, userID int64) (Game, error) {
	if gameID <= 0 {
		return Game{}, ErrInvalidGameID
	}
	if userID <= 0 {
		return Game{}, ErrInvalidUserID
	}

	updated, err := s.games.Join(ctx, gameID, userID)
	if err != nil {
		return Game{}, fmt.Errorf("join game: %w", err)
	}
	return updated, nil
}

// Leave removes the user from a game, except when the user is its organizer.
func (s *Service) Leave(ctx context.Context, gameID, userID int64) error {
	if gameID <= 0 {
		return ErrInvalidGameID
	}
	if userID <= 0 {
		return ErrInvalidUserID
	}

	if err := s.games.Leave(ctx, gameID, userID); err != nil {
		return fmt.Errorf("leave game: %w", err)
	}
	return nil
}

// SearchAvailable returns future games the user can join under current time rules.
func (s *Service) SearchAvailable(ctx context.Context, userID int64, filters SearchFilters) ([]SearchResult, error) {
	if userID <= 0 {
		return nil, ErrInvalidUserID
	}
	if (filters.DateStart == nil) != (filters.DateEnd == nil) ||
		(filters.DateStart != nil && !filters.DateEnd.After(*filters.DateStart)) {
		return nil, ErrInvalidSearchFilters
	}

	filters.City = strings.TrimSpace(filters.City)
	values, err := s.games.SearchAvailable(ctx, userID, filters)
	if err != nil {
		return nil, fmt.Errorf("search available games: %w", err)
	}
	if values == nil {
		values = []SearchResult{}
	}
	return values, nil
}
