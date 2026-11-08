package arena

import "context"

// Repository stores arenas without exposing database details to the domain.
type Repository interface {
	Create(ctx context.Context, value Arena) (Arena, error)
}
