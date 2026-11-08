package arena

import (
	"context"
	"fmt"
)

// Service applies arena registration rules and persistence operations.
type Service struct {
	repository Repository
}

// NewService creates an arena service with its persistence dependency.
func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

// Create validates the submitted details and assigns the authenticated owner.
func (s *Service) Create(ctx context.Context, ownerUserID int64, details Details) (Arena, error) {
	value, err := NewArena(ownerUserID, details)
	if err != nil {
		return Arena{}, err
	}

	createdArena, err := s.repository.Create(ctx, value)
	if err != nil {
		return Arena{}, fmt.Errorf("create arena: %w", err)
	}
	return createdArena, nil
}

// ListByOwner returns only arenas belonging to the given authenticated owner.
func (s *Service) ListByOwner(ctx context.Context, ownerUserID int64) ([]Arena, error) {
	if ownerUserID <= 0 {
		return nil, ErrInvalidOwner
	}

	arenas, err := s.repository.ListByOwner(ctx, ownerUserID)
	if err != nil {
		return nil, fmt.Errorf("list arenas: %w", err)
	}
	if arenas == nil {
		arenas = []Arena{}
	}
	return arenas, nil
}

// GetByIDAndOwner returns an arena only when it belongs to the requested owner.
func (s *Service) GetByIDAndOwner(ctx context.Context, arenaID, ownerUserID int64) (Arena, error) {
	if arenaID <= 0 {
		return Arena{}, ErrInvalidArenaID
	}
	if ownerUserID <= 0 {
		return Arena{}, ErrInvalidOwner
	}

	value, err := s.repository.GetByIDAndOwner(ctx, arenaID, ownerUserID)
	if err != nil {
		return Arena{}, fmt.Errorf("get arena by ID and owner: %w", err)
	}
	return value, nil
}

// Update validates and replaces the editable fields of an arena owned by the user.
func (s *Service) Update(ctx context.Context, arenaID, ownerUserID int64, details Details, active bool) (Arena, error) {
	if arenaID <= 0 {
		return Arena{}, ErrInvalidArenaID
	}

	value, err := NewArena(ownerUserID, details)
	if err != nil {
		return Arena{}, err
	}
	value.ID = arenaID
	value.Active = active

	updatedArena, err := s.repository.Update(ctx, value)
	if err != nil {
		return Arena{}, fmt.Errorf("update arena: %w", err)
	}
	return updatedArena, nil
}

// ToggleActive switches an owned arena between active and inactive.
func (s *Service) ToggleActive(ctx context.Context, arenaID, ownerUserID int64) (Arena, error) {
	if arenaID <= 0 {
		return Arena{}, ErrInvalidArenaID
	}
	if ownerUserID <= 0 {
		return Arena{}, ErrInvalidOwner
	}

	updatedArena, err := s.repository.ToggleActive(ctx, arenaID, ownerUserID)
	if err != nil {
		return Arena{}, fmt.Errorf("toggle arena active status: %w", err)
	}
	return updatedArena, nil
}
