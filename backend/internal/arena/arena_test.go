package arena

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewArenaNormalizesDetails(t *testing.T) {
	got, err := NewArena(17, Details{
		Name:      "  Arena Central  ",
		ArenaType: "  Society ",
		Address:   "  Rua do Esporte, 100 ",
		City:      "  São Paulo  ",
		State:     " sp ",
	})
	if err != nil {
		t.Fatalf("NewArena() error = %v, want nil", err)
	}
	if got.OwnerUserID != 17 || got.Name != "Arena Central" || got.ArenaType != "society" || got.Address != "Rua do Esporte, 100" || got.City != "São Paulo" || got.State != "SP" {
		t.Errorf("NewArena() = %+v, want normalized details and owner", got)
	}
}

func TestNewArenaRequiresOwner(t *testing.T) {
	_, err := NewArena(0, validDetails())
	if !errors.Is(err, ErrInvalidOwner) {
		t.Fatalf("NewArena() error = %v, want %v", err, ErrInvalidOwner)
	}
}

func TestNewArenaRequiresArenaDetails(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Details)
		wantErr error
	}{
		{name: "name", mutate: func(details *Details) { details.Name = " " }, wantErr: ErrInvalidName},
		{name: "type", mutate: func(details *Details) { details.ArenaType = " " }, wantErr: ErrInvalidType},
		{name: "address", mutate: func(details *Details) { details.Address = " " }, wantErr: ErrInvalidAddress},
		{name: "city", mutate: func(details *Details) { details.City = " " }, wantErr: ErrInvalidCity},
		{name: "state", mutate: func(details *Details) { details.State = " " }, wantErr: ErrInvalidState},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			details := validDetails()
			test.mutate(&details)
			if _, err := NewArena(17, details); !errors.Is(err, test.wantErr) {
				t.Fatalf("NewArena() error = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestNewArenaRejectsInvalidStateCode(t *testing.T) {
	for _, state := range []string{"S", "S1", "São"} {
		t.Run(state, func(t *testing.T) {
			details := validDetails()
			details.State = state
			if _, err := NewArena(17, details); !errors.Is(err, ErrInvalidState) {
				t.Errorf("NewArena() error = %v, want %v", err, ErrInvalidState)
			}
		})
	}
}

func TestServiceCreatePersistsValidArena(t *testing.T) {
	createdAt := time.Date(2026, time.November, 8, 12, 0, 0, 0, time.UTC)
	repository := &fakeRepository{createdArena: Arena{ID: 42, OwnerUserID: 17, CreatedAt: createdAt}}
	service := NewService(repository)

	created, err := service.Create(context.Background(), 17, validDetails())
	if err != nil {
		t.Fatalf("Create() error = %v, want nil", err)
	}
	if repository.createCalls != 1 || repository.savedArena.OwnerUserID != 17 {
		t.Fatalf("repository received %d creates and arena %+v", repository.createCalls, repository.savedArena)
	}
	if created.ID != 42 || !created.CreatedAt.Equal(createdAt) {
		t.Errorf("Create() = %+v, want persisted ID and timestamp", created)
	}
}

func TestServiceCreateReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	service := NewService(&fakeRepository{createError: wantErr})

	_, err := service.Create(context.Background(), 17, validDetails())
	if !errors.Is(err, wantErr) {
		t.Fatalf("Create() error = %v, want wrapped repository error %v", err, wantErr)
	}
}
func TestServiceCreateDoesNotPersistInvalidArena(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)
	details := validDetails()
	details.Name = " "

	_, err := service.Create(context.Background(), 17, details)
	if !errors.Is(err, ErrInvalidName) {
		t.Fatalf("Create() error = %v, want %v", err, ErrInvalidName)
	}
	if repository.createCalls != 0 {
		t.Errorf("repository create calls = %d, want 0", repository.createCalls)
	}
}

func validDetails() Details {
	return Details{
		Name:      "Arena Central",
		ArenaType: "society",
		Address:   "Rua do Esporte, 100",
		City:      "São Paulo",
		State:     "SP",
	}
}

type fakeRepository struct {
	createdArena Arena
	savedArena   Arena
	createCalls  int
	createError  error
}

func (r *fakeRepository) Create(_ context.Context, value Arena) (Arena, error) {
	r.createCalls++
	r.savedArena = value
	if r.createError != nil {
		return Arena{}, r.createError
	}
	created := r.createdArena
	created.OwnerUserID = value.OwnerUserID
	created.Name = value.Name
	created.ArenaType = value.ArenaType
	created.Address = value.Address
	created.City = value.City
	created.State = value.State
	return created, nil
}
