package game

import (
	"errors"
	"time"
)

var (
	ErrInvalidArenaID       = errors.New("arena ID is invalid")
	ErrInvalidTimeSlotID    = errors.New("time slot ID is invalid")
	ErrInvalidGameID        = errors.New("game ID is invalid")
	ErrInvalidUserID        = errors.New("user ID is invalid")
	ErrInvalidTimeRange     = errors.New("time slot end must be after its start")
	ErrTimeSlotInPast       = errors.New("time slot must start in the future")
	ErrTimeSlotConflict     = errors.New("time slot overlaps another slot in this arena")
	ErrTimeSlotUnavailable  = errors.New("time slot already has a game")
	ErrUserTimeConflict     = errors.New("user already has a game at that time")
	ErrAlreadyParticipating = errors.New("user already joined this game")
	ErrNotFound             = errors.New("game or time slot not found")
)

// TimeSlot is a bookable period configured for an arena.
type TimeSlot struct {
	ID        int64     `json:"id"`
	ArenaID   int64     `json:"arena_id"`
	StartsAt  time.Time `json:"starts_at"`
	EndsAt    time.Time `json:"ends_at"`
	CreatedAt time.Time `json:"created_at"`
}

// Game is a scheduled use of one arena time slot.
type Game struct {
	ID              int64     `json:"id"`
	TimeSlotID      int64     `json:"time_slot_id"`
	ArenaID         int64     `json:"arena_id"`
	OrganizerUserID int64     `json:"organizer_user_id"`
	StartsAt        time.Time `json:"starts_at"`
	EndsAt          time.Time `json:"ends_at"`
	Participants    []int64   `json:"participant_user_ids"`
	CreatedAt       time.Time `json:"created_at"`
}

// NewTimeSlot validates the arena and a future interval with a positive duration.
func NewTimeSlot(arenaID int64, startsAt, endsAt time.Time) (TimeSlot, error) {
	if arenaID <= 0 {
		return TimeSlot{}, ErrInvalidArenaID
	}
	if !startsAt.After(time.Now()) {
		return TimeSlot{}, ErrTimeSlotInPast
	}
	if !endsAt.After(startsAt) {
		return TimeSlot{}, ErrInvalidTimeRange
	}
	return TimeSlot{ArenaID: arenaID, StartsAt: startsAt, EndsAt: endsAt}, nil
}
