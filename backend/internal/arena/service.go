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
