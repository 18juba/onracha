package postgres

import (
	"context"
	"database/sql"
	"errors"
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

// ListByOwner returns arenas for one owner, newest first.
func (r *ArenaRepository) ListByOwner(ctx context.Context, ownerUserID int64) ([]arena.Arena, error) {
	const query = `
		SELECT id, owner_user_id, name, arena_type, address, city, state, created_at
		FROM arenas
		WHERE owner_user_id = $1
		ORDER BY created_at DESC, id DESC`

	rows, err := r.db.QueryContext(ctx, query, ownerUserID)
	if err != nil {
		return nil, fmt.Errorf("query arenas by owner: %w", err)
	}
	defer rows.Close()

	arenas := make([]arena.Arena, 0)
	for rows.Next() {
		var value arena.Arena
		if err := rows.Scan(
			&value.ID,
			&value.OwnerUserID,
			&value.Name,
			&value.ArenaType,
			&value.Address,
			&value.City,
			&value.State,
			&value.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan arena: %w", err)
		}
		arenas = append(arenas, value)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate arenas: %w", err)
	}
	return arenas, nil
}

// GetByIDAndOwner returns one arena only when it belongs to the supplied owner.
func (r *ArenaRepository) GetByIDAndOwner(ctx context.Context, arenaID, ownerUserID int64) (arena.Arena, error) {
	const query = `
		SELECT id, owner_user_id, name, arena_type, address, city, state, created_at
		FROM arenas
		WHERE id = $1 AND owner_user_id = $2`

	var value arena.Arena
	if err := r.db.QueryRowContext(ctx, query, arenaID, ownerUserID).Scan(
		&value.ID,
		&value.OwnerUserID,
		&value.Name,
		&value.ArenaType,
		&value.Address,
		&value.City,
		&value.State,
		&value.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return arena.Arena{}, arena.ErrNotFound
		}
		return arena.Arena{}, fmt.Errorf("query arena by ID and owner: %w", err)
	}
	return value, nil
}
