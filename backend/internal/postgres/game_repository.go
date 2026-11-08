package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"onracha/internal/game"
)

// GameRepository schedules games and tracks their participants in PostgreSQL.
type GameRepository struct {
	db *sql.DB
}

// NewGameRepository creates a game repository backed by PostgreSQL.
func NewGameRepository(db *sql.DB) *GameRepository {
	return &GameRepository{db: db}
}

// Schedule creates one game for a slot and enrolls its organizer as a participant.
func (r *GameRepository) Schedule(ctx context.Context, timeSlotID, organizerUserID int64) (game.Game, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return game.Game{}, fmt.Errorf("begin game scheduling: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := lockUser(ctx, tx, organizerUserID); err != nil {
		return game.Game{}, err
	}

	var interval game.TimeSlot
	if err := tx.QueryRowContext(ctx, `
		SELECT id, arena_id, starts_at, ends_at
		FROM arena_time_slots
		WHERE id = $1
		FOR UPDATE`, timeSlotID,
	).Scan(&interval.ID, &interval.ArenaID, &interval.StartsAt, &interval.EndsAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return game.Game{}, game.ErrNotFound
		}
		return game.Game{}, fmt.Errorf("lock game time slot: %w", err)
	}

	var alreadyScheduled bool
	if err := tx.QueryRowContext(ctx,
		"SELECT EXISTS (SELECT 1 FROM games WHERE time_slot_id = $1)", timeSlotID,
	).Scan(&alreadyScheduled); err != nil {
		return game.Game{}, fmt.Errorf("check time slot game: %w", err)
	}
	if alreadyScheduled {
		return game.Game{}, game.ErrTimeSlotUnavailable
	}

	conflict, err := userHasOverlappingGame(ctx, tx, organizerUserID, interval.StartsAt, interval.EndsAt, 0)
	if err != nil {
		return game.Game{}, fmt.Errorf("check organizer schedule: %w", err)
	}
	if conflict {
		return game.Game{}, game.ErrUserTimeConflict
	}

	var gameID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO games (time_slot_id, organizer_user_id)
		VALUES ($1, $2)
		RETURNING id`, timeSlotID, organizerUserID).Scan(&gameID); err != nil {
		return game.Game{}, fmt.Errorf("insert game: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO game_participants (game_id, user_id) VALUES ($1, $2)", gameID, organizerUserID,
	); err != nil {
		return game.Game{}, fmt.Errorf("enroll game organizer: %w", err)
	}

	created, err := loadGame(ctx, tx, gameID)
	if err != nil {
		return game.Game{}, err
	}
	if err := tx.Commit(); err != nil {
		return game.Game{}, fmt.Errorf("commit game scheduling: %w", err)
	}
	return created, nil
}

// ListByArena returns scheduled games with participant IDs in chronological order.
func (r *GameRepository) ListByArena(ctx context.Context, arenaID int64) ([]game.Game, error) {
	const query = `
		SELECT g.id, g.time_slot_id, slot.arena_id, g.organizer_user_id,
		       slot.starts_at, slot.ends_at, g.created_at, participant.user_id
		FROM games AS g
		JOIN arena_time_slots AS slot ON slot.id = g.time_slot_id
		LEFT JOIN game_participants AS participant ON participant.game_id = g.id
		WHERE slot.arena_id = $1
		ORDER BY slot.starts_at, g.id, participant.user_id`

	return scanGames(ctx, r.db, query, arenaID)
}

// Join enrolls the user when no other game overlaps the requested interval.
func (r *GameRepository) Join(ctx context.Context, gameID, userID int64) (game.Game, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return game.Game{}, fmt.Errorf("begin game participation: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := lockUser(ctx, tx, userID); err != nil {
		return game.Game{}, err
	}

	var interval game.TimeSlot
	if err := tx.QueryRowContext(ctx, `
		SELECT slot.id, slot.arena_id, slot.starts_at, slot.ends_at
		FROM games AS g
		JOIN arena_time_slots AS slot ON slot.id = g.time_slot_id
		WHERE g.id = $1
		FOR UPDATE OF g`, gameID,
	).Scan(&interval.ID, &interval.ArenaID, &interval.StartsAt, &interval.EndsAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return game.Game{}, game.ErrNotFound
		}
		return game.Game{}, fmt.Errorf("lock game for participation: %w", err)
	}

	var alreadyParticipating bool
	if err := tx.QueryRowContext(ctx,
		"SELECT EXISTS (SELECT 1 FROM game_participants WHERE game_id = $1 AND user_id = $2)", gameID, userID,
	).Scan(&alreadyParticipating); err != nil {
		return game.Game{}, fmt.Errorf("check existing game participation: %w", err)
	}
	if alreadyParticipating {
		return game.Game{}, game.ErrAlreadyParticipating
	}

	conflict, err := userHasOverlappingGame(ctx, tx, userID, interval.StartsAt, interval.EndsAt, gameID)
	if err != nil {
		return game.Game{}, fmt.Errorf("check participant schedule: %w", err)
	}
	if conflict {
		return game.Game{}, game.ErrUserTimeConflict
	}

	if _, err := tx.ExecContext(ctx,
		"INSERT INTO game_participants (game_id, user_id) VALUES ($1, $2)", gameID, userID,
	); err != nil {
		return game.Game{}, fmt.Errorf("add game participant: %w", err)
	}

	updated, err := loadGame(ctx, tx, gameID)
	if err != nil {
		return game.Game{}, err
	}
	if err := tx.Commit(); err != nil {
		return game.Game{}, fmt.Errorf("commit game participation: %w", err)
	}
	return updated, nil
}

func lockUser(ctx context.Context, tx *sql.Tx, userID int64) error {
	var lockedID int64
	if err := tx.QueryRowContext(ctx, "SELECT id FROM users WHERE id = $1 FOR UPDATE", userID).Scan(&lockedID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return game.ErrInvalidUserID
		}
		return fmt.Errorf("lock user schedule: %w", err)
	}
	return nil
}

func userHasOverlappingGame(ctx context.Context, tx *sql.Tx, userID int64, startsAt, endsAt time.Time, excludedGameID int64) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM game_participants AS participant
			JOIN games AS g ON g.id = participant.game_id
			JOIN arena_time_slots AS slot ON slot.id = g.time_slot_id
			WHERE participant.user_id = $1
			  AND g.id <> $4
			  AND slot.starts_at < $3
			  AND slot.ends_at > $2
		)`
	var conflict bool
	if err := tx.QueryRowContext(ctx, query, userID, startsAt, endsAt, excludedGameID).Scan(&conflict); err != nil {
		return false, err
	}
	return conflict, nil
}

func loadGame(ctx context.Context, tx *sql.Tx, gameID int64) (game.Game, error) {
	const query = `
		SELECT g.id, g.time_slot_id, slot.arena_id, g.organizer_user_id,
		       slot.starts_at, slot.ends_at, g.created_at, participant.user_id
		FROM games AS g
		JOIN arena_time_slots AS slot ON slot.id = g.time_slot_id
		LEFT JOIN game_participants AS participant ON participant.game_id = g.id
		WHERE g.id = $1
		ORDER BY participant.user_id`
	values, err := scanGames(ctx, tx, query, gameID)
	if err != nil {
		return game.Game{}, fmt.Errorf("load game: %w", err)
	}
	if len(values) == 0 {
		return game.Game{}, game.ErrNotFound
	}
	return values[0], nil
}

type gameQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func scanGames(ctx context.Context, queryer gameQueryer, query string, args ...any) ([]game.Game, error) {
	rows, err := queryer.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query scheduled games: %w", err)
	}
	defer rows.Close()

	values := make([]game.Game, 0)
	for rows.Next() {
		var value game.Game
		var participantID sql.NullInt64
		if err := rows.Scan(
			&value.ID,
			&value.TimeSlotID,
			&value.ArenaID,
			&value.OrganizerUserID,
			&value.StartsAt,
			&value.EndsAt,
			&value.CreatedAt,
			&participantID,
		); err != nil {
			return nil, fmt.Errorf("scan scheduled game: %w", err)
		}
		if len(values) == 0 || values[len(values)-1].ID != value.ID {
			value.Participants = make([]int64, 0)
			values = append(values, value)
		}
		if participantID.Valid {
			last := &values[len(values)-1]
			last.Participants = append(last.Participants, participantID.Int64)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate scheduled games: %w", err)
	}
	return values, nil
}
