package game

import (
	"context"
)

// TimeSlotRepository stores arena availability configured by its owner.
type TimeSlotRepository interface {
	Create(ctx context.Context, ownerUserID int64, value TimeSlot) (TimeSlot, error)
	ListAvailableByArena(ctx context.Context, arenaID int64) ([]TimeSlot, error)
}

// GameRepository schedules games and manages their participants.
type GameRepository interface {
	Schedule(ctx context.Context, timeSlotID, organizerUserID int64) (Game, error)
	ListByArena(ctx context.Context, arenaID int64) ([]Game, error)
	ListByParticipant(ctx context.Context, userID int64) ([]Game, error)
	GetByID(ctx context.Context, gameID int64) (Game, error)
	GetLobbyByID(ctx context.Context, gameID int64) (GameLobby, error)
	Join(ctx context.Context, gameID, userID int64) (Game, error)
	Leave(ctx context.Context, gameID, userID int64) error
	SearchAvailable(ctx context.Context, userID int64, filters SearchFilters) ([]SearchResult, error)
}
