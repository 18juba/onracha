package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"onracha/internal/game"
)

// TimeSlotRepository stores availability configured for arenas.
type TimeSlotRepository struct {
	db *sql.DB
}

// NewTimeSlotRepository creates a time slot repository backed by PostgreSQL.
func NewTimeSlotRepository(db *sql.DB) *TimeSlotRepository {
	return &TimeSlotRepository{db: db}
}

// Create adds an interval when the user owns the arena and it does not overlap another slot.
func (r *TimeSlotRepository) Create(ctx context.Context, ownerUserID int64, value game.TimeSlot) (game.TimeSlot, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return game.TimeSlot{}, fmt.Errorf("begin time slot creation: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var arenaID int64
	if err := tx.QueryRowContext(ctx,
		"SELECT id FROM arenas WHERE id = $1 AND owner_user_id = $2 FOR UPDATE",
		value.ArenaID, ownerUserID,
	).Scan(&arenaID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return game.TimeSlot{}, game.ErrNotFound
		}
		return game.TimeSlot{}, fmt.Errorf("lock arena for time slot: %w", err)
	}

	var overlaps bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM arena_time_slots
			WHERE arena_id = $1 AND starts_at < $3 AND ends_at > $2
		)`, value.ArenaID, value.StartsAt, value.EndsAt).Scan(&overlaps); err != nil {
		return game.TimeSlot{}, fmt.Errorf("check overlapping arena time slots: %w", err)
	}
	if overlaps {
		return game.TimeSlot{}, game.ErrTimeSlotConflict
	}

	if err := tx.QueryRowContext(ctx, `
		INSERT INTO arena_time_slots (arena_id, starts_at, ends_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`, value.ArenaID, value.StartsAt, value.EndsAt,
	).Scan(&value.ID, &value.CreatedAt); err != nil {
		return game.TimeSlot{}, fmt.Errorf("insert arena time slot: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return game.TimeSlot{}, fmt.Errorf("commit time slot creation: %w", err)
	}
	return value, nil
}

// ListAvailableByArena returns future intervals that do not have a scheduled game.
func (r *TimeSlotRepository) ListAvailableByArena(ctx context.Context, arenaID int64) ([]game.TimeSlot, error) {
	const query = `
		SELECT slot.id, slot.arena_id, slot.starts_at, slot.ends_at, slot.created_at
		FROM arena_time_slots AS slot
		WHERE slot.arena_id = $1
		  AND slot.starts_at > NOW()
		  AND NOT EXISTS (
			SELECT 1 FROM games WHERE games.time_slot_id = slot.id
		  )
		ORDER BY slot.starts_at, slot.id`

	rows, err := r.db.QueryContext(ctx, query, arenaID)
	if err != nil {
		return nil, fmt.Errorf("query available time slots: %w", err)
	}
	defer rows.Close()

	values := make([]game.TimeSlot, 0)
	for rows.Next() {
		var value game.TimeSlot
		if err := rows.Scan(&value.ID, &value.ArenaID, &value.StartsAt, &value.EndsAt, &value.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan available time slot: %w", err)
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate available time slots: %w", err)
	}
	return values, nil
}
