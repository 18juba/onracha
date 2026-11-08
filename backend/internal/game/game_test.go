package game

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewTimeSlotBuildsFutureInterval(t *testing.T) {
	startsAt := time.Now().Add(2 * time.Hour).Truncate(time.Second)
	endsAt := startsAt.Add(time.Hour)

	got, err := NewTimeSlot(23, startsAt, endsAt)
	if err != nil {
		t.Fatalf("NewTimeSlot() error = %v, want nil", err)
	}
	if got.ArenaID != 23 || !got.StartsAt.Equal(startsAt) || !got.EndsAt.Equal(endsAt) {
		t.Errorf("NewTimeSlot() = %+v, want arena 23 from %v to %v", got, startsAt, endsAt)
	}
}

func TestNewTimeSlotRejectsInvalidArena(t *testing.T) {
	startsAt := time.Now().Add(time.Hour)
	_, err := NewTimeSlot(0, startsAt, startsAt.Add(time.Hour))
	if !errors.Is(err, ErrInvalidArenaID) {
		t.Fatalf("NewTimeSlot() error = %v, want %v", err, ErrInvalidArenaID)
	}
}

func TestNewTimeSlotRejectsPastStart(t *testing.T) {
	startsAt := time.Now().Add(-time.Hour)
	_, err := NewTimeSlot(23, startsAt, startsAt.Add(time.Hour))
	if !errors.Is(err, ErrTimeSlotInPast) {
		t.Fatalf("NewTimeSlot() error = %v, want %v", err, ErrTimeSlotInPast)
	}
}

func TestNewTimeSlotRequiresEndAfterStart(t *testing.T) {
	startsAt := time.Now().Add(time.Hour)
	for _, endsAt := range []time.Time{startsAt, startsAt.Add(-time.Minute)} {
		if _, err := NewTimeSlot(23, startsAt, endsAt); !errors.Is(err, ErrInvalidTimeRange) {
			t.Errorf("NewTimeSlot(%v) error = %v, want %v", endsAt, err, ErrInvalidTimeRange)
		}
	}
}

func TestServiceCreateTimeSlotUsesOwnerRepository(t *testing.T) {
	startsAt := time.Now().Add(2 * time.Hour).Truncate(time.Second)
	endsAt := startsAt.Add(time.Hour)
	want := TimeSlot{ID: 9, ArenaID: 23, StartsAt: startsAt, EndsAt: endsAt}
	repository := &fakeTimeSlotRepository{createdSlot: want}
	service := NewService(repository, &fakeGameRepository{})

	got, err := service.CreateTimeSlot(context.Background(), 23, 17, startsAt, endsAt)
	if err != nil {
		t.Fatalf("CreateTimeSlot() error = %v, want nil", err)
	}
	if repository.createCalls != 1 || repository.ownerID != 17 || repository.created.ArenaID != 23 {
		t.Errorf("repository create calls = %d, owner ID = %d, slot = %+v", repository.createCalls, repository.ownerID, repository.created)
	}
	if got.ID != want.ID || !got.StartsAt.Equal(want.StartsAt) || !got.EndsAt.Equal(want.EndsAt) {
		t.Errorf("CreateTimeSlot() = %+v, want %+v", got, want)
	}
}

func TestServiceCreateTimeSlotRejectsInvalidOwner(t *testing.T) {
	repository := &fakeTimeSlotRepository{}
	service := NewService(repository, &fakeGameRepository{})
	startsAt := time.Now().Add(time.Hour)

	_, err := service.CreateTimeSlot(context.Background(), 23, 0, startsAt, startsAt.Add(time.Hour))
	if !errors.Is(err, ErrInvalidUserID) {
		t.Fatalf("CreateTimeSlot() error = %v, want %v", err, ErrInvalidUserID)
	}
	if repository.createCalls != 0 {
		t.Errorf("repository create calls = %d, want 0", repository.createCalls)
	}
}

func TestServiceCreateTimeSlotReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	repository := &fakeTimeSlotRepository{createError: wantErr}
	service := NewService(repository, &fakeGameRepository{})
	startsAt := time.Now().Add(time.Hour)

	_, err := service.CreateTimeSlot(context.Background(), 23, 17, startsAt, startsAt.Add(time.Hour))
	if !errors.Is(err, wantErr) {
		t.Fatalf("CreateTimeSlot() error = %v, want wrapped error %v", err, wantErr)
	}
}

func TestServiceListAvailableTimeSlotsRejectsInvalidArena(t *testing.T) {
	repository := &fakeTimeSlotRepository{}
	service := NewService(repository, &fakeGameRepository{})

	_, err := service.ListAvailableTimeSlots(context.Background(), 0)
	if !errors.Is(err, ErrInvalidArenaID) {
		t.Fatalf("ListAvailableTimeSlots() error = %v, want %v", err, ErrInvalidArenaID)
	}
	if repository.listCalls != 0 {
		t.Errorf("repository list calls = %d, want 0", repository.listCalls)
	}
}

func TestServiceListAvailableTimeSlotsReturnsRepositoryValues(t *testing.T) {
	want := []TimeSlot{{ID: 9, ArenaID: 23}}
	repository := &fakeTimeSlotRepository{listedSlots: want}
	service := NewService(repository, &fakeGameRepository{})

	got, err := service.ListAvailableTimeSlots(context.Background(), 23)
	if err != nil {
		t.Fatalf("ListAvailableTimeSlots() error = %v, want nil", err)
	}
	if repository.listCalls != 1 || repository.listedArenaID != 23 || len(got) != 1 || got[0].ID != want[0].ID {
		t.Errorf("ListAvailableTimeSlots() = %+v; repository calls=%d arena=%d, want slot %+v", got, repository.listCalls, repository.listedArenaID, want)
	}
}

func TestServiceListAvailableTimeSlotsReturnsEmptySlice(t *testing.T) {
	service := NewService(&fakeTimeSlotRepository{}, &fakeGameRepository{})

	got, err := service.ListAvailableTimeSlots(context.Background(), 23)
	if err != nil {
		t.Fatalf("ListAvailableTimeSlots() error = %v, want nil", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("ListAvailableTimeSlots() = %+v, want non-nil empty slice", got)
	}
}

func TestServiceListAvailableTimeSlotsReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	service := NewService(&fakeTimeSlotRepository{listError: wantErr}, &fakeGameRepository{})

	_, err := service.ListAvailableTimeSlots(context.Background(), 23)
	if !errors.Is(err, wantErr) {
		t.Fatalf("ListAvailableTimeSlots() error = %v, want wrapped error %v", err, wantErr)
	}
}

func TestServiceSchedulePassesSlotAndOrganizer(t *testing.T) {
	want := Game{ID: 31, TimeSlotID: 9, OrganizerUserID: 17}
	repository := &fakeGameRepository{scheduledGame: want}
	service := NewService(&fakeTimeSlotRepository{}, repository)

	got, err := service.Schedule(context.Background(), 9, 17)
	if err != nil {
		t.Fatalf("Schedule() error = %v, want nil", err)
	}
	if repository.scheduleCalls != 1 || repository.slotID != 9 || repository.userID != 17 {
		t.Errorf("schedule calls = %d, slot ID = %d, user ID = %d", repository.scheduleCalls, repository.slotID, repository.userID)
	}
	if got.ID != want.ID || got.TimeSlotID != want.TimeSlotID {
		t.Errorf("Schedule() = %+v, want %+v", got, want)
	}
}

func TestServiceScheduleRejectsInvalidIdentifiers(t *testing.T) {
	tests := []struct {
		name   string
		slotID int64
		userID int64
		want   error
	}{
		{name: "slot", slotID: 0, userID: 17, want: ErrInvalidTimeSlotID},
		{name: "user", slotID: 9, userID: 0, want: ErrInvalidUserID},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeGameRepository{}
			service := NewService(&fakeTimeSlotRepository{}, repository)
			if _, err := service.Schedule(context.Background(), test.slotID, test.userID); !errors.Is(err, test.want) {
				t.Fatalf("Schedule() error = %v, want %v", err, test.want)
			}
			if repository.scheduleCalls != 0 {
				t.Errorf("schedule calls = %d, want 0", repository.scheduleCalls)
			}
		})
	}
}

func TestServiceScheduleReturnsUserTimeConflict(t *testing.T) {
	service := NewService(&fakeTimeSlotRepository{}, &fakeGameRepository{scheduleError: ErrUserTimeConflict})

	_, err := service.Schedule(context.Background(), 9, 17)
	if !errors.Is(err, ErrUserTimeConflict) {
		t.Fatalf("Schedule() error = %v, want %v", err, ErrUserTimeConflict)
	}
}

func TestServiceListByArenaReturnsRepositoryGames(t *testing.T) {
	want := []Game{{ID: 31, ArenaID: 23}}
	repository := &fakeGameRepository{listedGames: want}
	service := NewService(&fakeTimeSlotRepository{}, repository)

	got, err := service.ListByArena(context.Background(), 23)
	if err != nil {
		t.Fatalf("ListByArena() error = %v, want nil", err)
	}
	if repository.listCalls != 1 || repository.listedArenaID != 23 || len(got) != 1 || got[0].ID != want[0].ID {
		t.Errorf("ListByArena() = %+v; repository calls=%d arena=%d, want game %+v", got, repository.listCalls, repository.listedArenaID, want)
	}
}

func TestServiceListByArenaReturnsEmptySlice(t *testing.T) {
	service := NewService(&fakeTimeSlotRepository{}, &fakeGameRepository{})

	got, err := service.ListByArena(context.Background(), 23)
	if err != nil {
		t.Fatalf("ListByArena() error = %v, want nil", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("ListByArena() = %+v, want non-nil empty slice", got)
	}
}

func TestServiceListByArenaRejectsInvalidArenaID(t *testing.T) {
	repository := &fakeGameRepository{}
	service := NewService(&fakeTimeSlotRepository{}, repository)

	_, err := service.ListByArena(context.Background(), 0)
	if !errors.Is(err, ErrInvalidArenaID) {
		t.Fatalf("ListByArena() error = %v, want %v", err, ErrInvalidArenaID)
	}
	if repository.listCalls != 0 {
		t.Errorf("repository list calls = %d, want 0", repository.listCalls)
	}
}

func TestServiceListByArenaReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	service := NewService(&fakeTimeSlotRepository{}, &fakeGameRepository{listError: wantErr})

	_, err := service.ListByArena(context.Background(), 23)
	if !errors.Is(err, wantErr) {
		t.Fatalf("ListByArena() error = %v, want wrapped error %v", err, wantErr)
	}
}

func TestServiceListByParticipantReturnsRepositoryGames(t *testing.T) {
	want := []Game{{ID: 31, OrganizerUserID: 17}}
	repository := &fakeGameRepository{participantGames: want}
	service := NewService(&fakeTimeSlotRepository{}, repository)

	got, err := service.ListByParticipant(context.Background(), 22)
	if err != nil {
		t.Fatalf("ListByParticipant() error = %v, want nil", err)
	}
	if repository.participantListCalls != 1 || repository.participantUserID != 22 || len(got) != 1 || got[0].ID != want[0].ID {
		t.Errorf("ListByParticipant() = %+v; repository calls=%d user=%d, want games %+v", got, repository.participantListCalls, repository.participantUserID, want)
	}
}

func TestServiceListByParticipantReturnsEmptySlice(t *testing.T) {
	service := NewService(&fakeTimeSlotRepository{}, &fakeGameRepository{})

	got, err := service.ListByParticipant(context.Background(), 22)
	if err != nil {
		t.Fatalf("ListByParticipant() error = %v, want nil", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("ListByParticipant() = %+v, want non-nil empty slice", got)
	}
}

func TestServiceListByParticipantRejectsInvalidUserID(t *testing.T) {
	repository := &fakeGameRepository{}
	service := NewService(&fakeTimeSlotRepository{}, repository)

	if _, err := service.ListByParticipant(context.Background(), 0); !errors.Is(err, ErrInvalidUserID) {
		t.Fatalf("ListByParticipant() error = %v, want %v", err, ErrInvalidUserID)
	}
	if repository.participantListCalls != 0 {
		t.Errorf("repository calls = %d, want 0", repository.participantListCalls)
	}
}

func TestServiceListByParticipantReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	service := NewService(&fakeTimeSlotRepository{}, &fakeGameRepository{participantListError: wantErr})

	if _, err := service.ListByParticipant(context.Background(), 22); !errors.Is(err, wantErr) {
		t.Fatalf("ListByParticipant() error = %v, want wrapped error %v", err, wantErr)
	}
}

func TestServiceGetByIDReturnsGame(t *testing.T) {
	want := Game{ID: 31, ArenaID: 23, Participants: []int64{17, 22}}
	repository := &fakeGameRepository{gameByID: want}
	service := NewService(&fakeTimeSlotRepository{}, repository)

	got, err := service.GetByID(context.Background(), 31)
	if err != nil {
		t.Fatalf("GetByID() error = %v, want nil", err)
	}
	if repository.getByIDCalls != 1 || repository.requestedGameID != 31 || got.ID != want.ID || len(got.Participants) != 2 {
		t.Errorf("GetByID() = %+v; repository calls=%d requested ID=%d, want %+v", got, repository.getByIDCalls, repository.requestedGameID, want)
	}
}

func TestServiceGetByIDRejectsInvalidID(t *testing.T) {
	repository := &fakeGameRepository{}
	service := NewService(&fakeTimeSlotRepository{}, repository)

	if _, err := service.GetByID(context.Background(), 0); !errors.Is(err, ErrInvalidGameID) {
		t.Fatalf("GetByID() error = %v, want %v", err, ErrInvalidGameID)
	}
	if repository.getByIDCalls != 0 {
		t.Errorf("repository calls = %d, want 0", repository.getByIDCalls)
	}
}

func TestServiceGetByIDReturnsRepositoryError(t *testing.T) {
	wantErr := ErrNotFound
	service := NewService(&fakeTimeSlotRepository{}, &fakeGameRepository{getByIDError: wantErr})

	if _, err := service.GetByID(context.Background(), 31); !errors.Is(err, wantErr) {
		t.Fatalf("GetByID() error = %v, want wrapped error %v", err, wantErr)
	}
}

func TestServiceJoinPassesGameAndUser(t *testing.T) {
	want := Game{ID: 31, Participants: []int64{17, 22}}
	repository := &fakeGameRepository{joinedGame: want}
	service := NewService(&fakeTimeSlotRepository{}, repository)

	got, err := service.Join(context.Background(), 31, 22)
	if err != nil {
		t.Fatalf("Join() error = %v, want nil", err)
	}
	if repository.joinCalls != 1 || repository.gameID != 31 || repository.userID != 22 {
		t.Errorf("join calls = %d, game ID = %d, user ID = %d", repository.joinCalls, repository.gameID, repository.userID)
	}
	if got.ID != want.ID || len(got.Participants) != 2 {
		t.Errorf("Join() = %+v, want %+v", got, want)
	}
}

func TestServiceJoinRejectsInvalidIdentifiers(t *testing.T) {
	tests := []struct {
		name   string
		gameID int64
		userID int64
		want   error
	}{
		{name: "game", gameID: 0, userID: 17, want: ErrInvalidGameID},
		{name: "user", gameID: 31, userID: 0, want: ErrInvalidUserID},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeGameRepository{}
			service := NewService(&fakeTimeSlotRepository{}, repository)
			if _, err := service.Join(context.Background(), test.gameID, test.userID); !errors.Is(err, test.want) {
				t.Fatalf("Join() error = %v, want %v", err, test.want)
			}
			if repository.joinCalls != 0 {
				t.Errorf("join calls = %d, want 0", repository.joinCalls)
			}
		})
	}
}

func TestServiceJoinReturnsRepositoryError(t *testing.T) {
	service := NewService(&fakeTimeSlotRepository{}, &fakeGameRepository{joinError: ErrAlreadyParticipating})

	_, err := service.Join(context.Background(), 31, 22)
	if !errors.Is(err, ErrAlreadyParticipating) {
		t.Fatalf("Join() error = %v, want %v", err, ErrAlreadyParticipating)
	}
}

func TestServiceLeavePassesGameAndUser(t *testing.T) {
	repository := &fakeGameRepository{}
	service := NewService(&fakeTimeSlotRepository{}, repository)

	if err := service.Leave(context.Background(), 31, 22); err != nil {
		t.Fatalf("Leave() error = %v, want nil", err)
	}
	if repository.leaveCalls != 1 || repository.leaveGameID != 31 || repository.leaveUserID != 22 {
		t.Errorf("leave calls = %d, game ID = %d, user ID = %d; want 1, 31, 22", repository.leaveCalls, repository.leaveGameID, repository.leaveUserID)
	}
}

func TestServiceLeaveRejectsInvalidIdentifiers(t *testing.T) {
	tests := []struct {
		name   string
		gameID int64
		userID int64
		want   error
	}{
		{name: "game", gameID: 0, userID: 22, want: ErrInvalidGameID},
		{name: "user", gameID: 31, userID: 0, want: ErrInvalidUserID},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeGameRepository{}
			service := NewService(&fakeTimeSlotRepository{}, repository)
			if err := service.Leave(context.Background(), test.gameID, test.userID); !errors.Is(err, test.want) {
				t.Fatalf("Leave() error = %v, want %v", err, test.want)
			}
			if repository.leaveCalls != 0 {
				t.Errorf("repository calls = %d, want 0", repository.leaveCalls)
			}
		})
	}
}

func TestServiceLeaveReturnsRepositoryError(t *testing.T) {
	wantErr := ErrOrganizerCannotLeave
	service := NewService(&fakeTimeSlotRepository{}, &fakeGameRepository{leaveError: wantErr})

	if err := service.Leave(context.Background(), 31, 22); !errors.Is(err, wantErr) {
		t.Fatalf("Leave() error = %v, want wrapped error %v", err, wantErr)
	}
}

func TestServiceSearchAvailablePassesNormalizedFilters(t *testing.T) {
	startsAt := time.Date(2026, time.December, 12, 0, 0, 0, 0, time.UTC)
	endsAt := startsAt.Add(24 * time.Hour)
	want := []SearchResult{{Game: Game{ID: 31}, Arena: ArenaSummary{ID: 23, City: "São Paulo"}}}
	repository := &fakeGameRepository{searchResults: want}
	service := NewService(&fakeTimeSlotRepository{}, repository)

	got, err := service.SearchAvailable(context.Background(), 22, SearchFilters{
		City: " São Paulo ", DateStart: &startsAt, DateEnd: &endsAt,
	})
	if err != nil {
		t.Fatalf("SearchAvailable() error = %v, want nil", err)
	}
	if repository.searchCalls != 1 || repository.searchUserID != 22 || repository.searchFilters.City != "São Paulo" ||
		!repository.searchFilters.DateStart.Equal(startsAt) || !repository.searchFilters.DateEnd.Equal(endsAt) ||
		len(got) != 1 || got[0].Game.ID != want[0].Game.ID {
		t.Errorf("SearchAvailable() = %+v; repository calls=%d user=%d filters=%+v, want %+v", got, repository.searchCalls, repository.searchUserID, repository.searchFilters, want)
	}
}

func TestServiceSearchAvailableReturnsEmptySlice(t *testing.T) {
	service := NewService(&fakeTimeSlotRepository{}, &fakeGameRepository{})

	got, err := service.SearchAvailable(context.Background(), 22, SearchFilters{})
	if err != nil {
		t.Fatalf("SearchAvailable() error = %v, want nil", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("SearchAvailable() = %+v, want non-nil empty slice", got)
	}
}

func TestServiceSearchAvailableRejectsInvalidUserOrDateRange(t *testing.T) {
	startsAt := time.Date(2026, time.December, 12, 0, 0, 0, 0, time.UTC)
	endsAt := startsAt
	tests := []struct {
		name    string
		userID  int64
		filters SearchFilters
		want    error
	}{
		{name: "user", userID: 0, want: ErrInvalidUserID},
		{name: "missing end date", userID: 22, filters: SearchFilters{DateStart: &startsAt}, want: ErrInvalidSearchFilters},
		{name: "empty date range", userID: 22, filters: SearchFilters{DateStart: &startsAt, DateEnd: &endsAt}, want: ErrInvalidSearchFilters},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeGameRepository{}
			service := NewService(&fakeTimeSlotRepository{}, repository)
			if _, err := service.SearchAvailable(context.Background(), test.userID, test.filters); !errors.Is(err, test.want) {
				t.Fatalf("SearchAvailable() error = %v, want %v", err, test.want)
			}
			if repository.searchCalls != 0 {
				t.Errorf("repository calls = %d, want 0", repository.searchCalls)
			}
		})
	}
}

func TestServiceSearchAvailableReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	service := NewService(&fakeTimeSlotRepository{}, &fakeGameRepository{searchError: wantErr})

	if _, err := service.SearchAvailable(context.Background(), 22, SearchFilters{}); !errors.Is(err, wantErr) {
		t.Fatalf("SearchAvailable() error = %v, want wrapped error %v", err, wantErr)
	}
}

type fakeTimeSlotRepository struct {
	createdSlot   TimeSlot
	created       TimeSlot
	ownerID       int64
	createCalls   int
	createError   error
	listCalls     int
	listedArenaID int64
	listedSlots   []TimeSlot
	listError     error
}

func (r *fakeTimeSlotRepository) Create(_ context.Context, ownerUserID int64, value TimeSlot) (TimeSlot, error) {
	r.createCalls++
	r.ownerID = ownerUserID
	r.created = value
	if r.createError != nil {
		return TimeSlot{}, r.createError
	}
	return r.createdSlot, nil
}

func (r *fakeTimeSlotRepository) ListAvailableByArena(_ context.Context, arenaID int64) ([]TimeSlot, error) {
	r.listCalls++
	r.listedArenaID = arenaID
	return r.listedSlots, r.listError
}

type fakeGameRepository struct {
	scheduledGame        Game
	scheduleCalls        int
	slotID               int64
	userID               int64
	scheduleError        error
	joinedGame           Game
	joinCalls            int
	gameID               int64
	joinError            error
	listCalls            int
	listedArenaID        int64
	listedGames          []Game
	listError            error
	participantListCalls int
	participantUserID    int64
	participantGames     []Game
	participantListError error
	getByIDCalls         int
	requestedGameID      int64
	gameByID             Game
	getByIDError         error
	leaveCalls           int
	leaveGameID          int64
	leaveUserID          int64
	leaveError           error
	searchCalls          int
	searchUserID         int64
	searchFilters        SearchFilters
	searchResults        []SearchResult
	searchError          error
}

func (r *fakeGameRepository) Schedule(_ context.Context, slotID, organizerUserID int64) (Game, error) {
	r.scheduleCalls++
	r.slotID = slotID
	r.userID = organizerUserID
	if r.scheduleError != nil {
		return Game{}, r.scheduleError
	}
	return r.scheduledGame, nil
}

func (r *fakeGameRepository) ListByArena(_ context.Context, arenaID int64) ([]Game, error) {
	r.listCalls++
	r.listedArenaID = arenaID
	return r.listedGames, r.listError
}

func (r *fakeGameRepository) ListByParticipant(_ context.Context, userID int64) ([]Game, error) {
	r.participantListCalls++
	r.participantUserID = userID
	return r.participantGames, r.participantListError
}

func (r *fakeGameRepository) GetByID(_ context.Context, gameID int64) (Game, error) {
	r.getByIDCalls++
	r.requestedGameID = gameID
	return r.gameByID, r.getByIDError
}

func (r *fakeGameRepository) Join(_ context.Context, gameID, userID int64) (Game, error) {
	r.joinCalls++
	r.gameID = gameID
	r.userID = userID
	if r.joinError != nil {
		return Game{}, r.joinError
	}
	return r.joinedGame, nil
}

func (r *fakeGameRepository) Leave(_ context.Context, gameID, userID int64) error {
	r.leaveCalls++
	r.leaveGameID = gameID
	r.leaveUserID = userID
	return r.leaveError
}

func (r *fakeGameRepository) SearchAvailable(_ context.Context, userID int64, filters SearchFilters) ([]SearchResult, error) {
	r.searchCalls++
	r.searchUserID = userID
	r.searchFilters = filters
	return r.searchResults, r.searchError
}
