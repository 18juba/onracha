package arena

import "context"

// Repository stores arenas without exposing database details to the domain.
type Repository interface {
	Create(ctx context.Context, value Arena) (Arena, error)
	ListByOwner(ctx context.Context, ownerUserID int64) ([]Arena, error)
	GetByIDAndOwner(ctx context.Context, arenaID, ownerUserID int64) (Arena, error)
	Update(ctx context.Context, value Arena) (Arena, error)
	ToggleActive(ctx context.Context, arenaID, ownerUserID int64) (Arena, error)
}
