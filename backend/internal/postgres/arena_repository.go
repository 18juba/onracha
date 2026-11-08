package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"onracha/internal/arena"
)

// ArenaRepository persists arena records in PostgreSQL.
type ArenaRepository struct {
	db *sql.DB
}

// NewArenaRepository creates an arena repository backed by an existing connection pool.
func NewArenaRepository(db *sql.DB) *ArenaRepository {
	return &ArenaRepository{db: db}
}

// Create inserts an arena owned by the authenticated user and returns its database fields.
func (r *ArenaRepository) Create(ctx context.Context, value arena.Arena) (arena.Arena, error) {
	const query = `
		INSERT INTO arenas (owner_user_id, name, arena_type, address, city, state)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`
	if err := r.db.QueryRowContext(ctx, query,
		value.OwnerUserID,
		value.Name,
		value.ArenaType,
		value.Address,
		value.City,
		value.State,
	).Scan(&value.ID, &value.CreatedAt); err != nil {
		return arena.Arena{}, fmt.Errorf("insert arena: %w", err)
	}
	return value, nil
}
