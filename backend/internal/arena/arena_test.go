package arena

import (
	"context"
	"errors"
	"math"
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
	if !got.Active {
		t.Error("NewArena().Active = false, want newly registered arenas to be active")
	}
}

func TestNewArenaKeepsCoordinates(t *testing.T) {
	latitude := -23.55052
	longitude := -46.633308
	details := validDetails()
	details.Latitude = &latitude
	details.Longitude = &longitude

	got, err := NewArena(17, details)
	if err != nil {
		t.Fatalf("NewArena() error = %v, want nil", err)
	}
	if got.Latitude == nil || *got.Latitude != latitude || got.Longitude == nil || *got.Longitude != longitude {
		t.Errorf("NewArena() coordinates = (%v, %v), want (%v, %v)", got.Latitude, got.Longitude, latitude, longitude)
	}
}

func TestNewArenaAllowsMissingCoordinates(t *testing.T) {
	got, err := NewArena(17, validDetails())
	if err != nil {
		t.Fatalf("NewArena() error = %v, want nil", err)
	}
	if got.Latitude != nil || got.Longitude != nil {
		t.Errorf("NewArena() optional coordinates = (%v, %v), want nil values", got.Latitude, got.Longitude)
	}
}

func TestNewArenaRequiresLatitudeAndLongitudeTogether(t *testing.T) {
	latitude := -23.55052
	longitude := -46.633308
	tests := []struct {
		name      string
		latitude  *float64
		longitude *float64
	}{
		{name: "latitude only", latitude: &latitude},
		{name: "longitude only", longitude: &longitude},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			details := validDetails()
			details.Latitude = test.latitude
			details.Longitude = test.longitude
			if _, err := NewArena(17, details); !errors.Is(err, ErrInvalidLocation) {
				t.Errorf("NewArena() error = %v, want %v", err, ErrInvalidLocation)
			}
		})
	}
}

func TestNewArenaRejectsCoordinatesOutsideGeographicBounds(t *testing.T) {
	tests := []struct {
		name      string
		latitude  float64
		longitude float64
	}{
		{name: "latitude below minimum", latitude: -90.01, longitude: 0},
		{name: "latitude above maximum", latitude: 90.01, longitude: 0},
		{name: "longitude below minimum", latitude: 0, longitude: -180.01},
		{name: "longitude above maximum", latitude: 0, longitude: 180.01},
		{name: "not a number", latitude: math.NaN(), longitude: 0},
		{name: "infinite coordinate", latitude: 0, longitude: math.Inf(1)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			details := validDetails()
			details.Latitude = &test.latitude
			details.Longitude = &test.longitude
			if _, err := NewArena(17, details); !errors.Is(err, ErrInvalidLocation) {
				t.Errorf("NewArena() error = %v, want %v", err, ErrInvalidLocation)
			}
		})
	}
}

func TestNewArenaAcceptsCoordinateBoundaries(t *testing.T) {
	latitude := -90.0
	longitude := 180.0
	details := validDetails()
	details.Latitude = &latitude
	details.Longitude = &longitude

	if _, err := NewArena(17, details); err != nil {
		t.Errorf("NewArena() at valid geographic boundaries error = %v, want nil", err)
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
	if created.ID != 42 || !created.CreatedAt.Equal(createdAt) || !created.Active {
		t.Errorf("Create() = %+v, want persisted ID, timestamp and active status", created)
	}
}

func TestServiceListByOwnerReturnsRepositoryResults(t *testing.T) {
	want := []Arena{{ID: 7, OwnerUserID: 17, Name: "Arena Central"}}
	repository := &fakeRepository{listedArenas: want}
	service := NewService(repository)

	got, err := service.ListByOwner(context.Background(), 17)
	if err != nil {
		t.Fatalf("ListByOwner() error = %v, want nil", err)
	}
	if repository.listCalls != 1 || repository.listedOwnerID != 17 {
		t.Errorf("repository list calls = %d, owner ID = %d; want one call for 17", repository.listCalls, repository.listedOwnerID)
	}
	if len(got) != 1 || got[0].ID != want[0].ID {
		t.Errorf("ListByOwner() = %+v, want %+v", got, want)
	}
}

func TestServiceListByOwnerReturnsEmptySliceWhenNoArenas(t *testing.T) {
	service := NewService(&fakeRepository{})

	got, err := service.ListByOwner(context.Background(), 17)
	if err != nil {
		t.Fatalf("ListByOwner() error = %v, want nil", err)
	}
	if got == nil || len(got) != 0 {
		t.Errorf("ListByOwner() = %+v, want non-nil empty slice", got)
	}
}

func TestServiceListByOwnerRejectsInvalidOwner(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	_, err := service.ListByOwner(context.Background(), 0)
	if !errors.Is(err, ErrInvalidOwner) {
		t.Fatalf("ListByOwner() error = %v, want %v", err, ErrInvalidOwner)
	}
	if repository.listCalls != 0 {
		t.Errorf("repository list calls = %d, want 0", repository.listCalls)
	}
}

func TestServiceListByOwnerReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	service := NewService(&fakeRepository{listError: wantErr})

	_, err := service.ListByOwner(context.Background(), 17)
	if !errors.Is(err, wantErr) {
		t.Fatalf("ListByOwner() error = %v, want wrapped repository error %v", err, wantErr)
	}
}

func TestServiceGetByIDAndOwnerReturnsRepositoryArena(t *testing.T) {
	want := Arena{ID: 42, OwnerUserID: 17, Name: "Arena Central"}
	repository := &fakeRepository{queriedArena: want}
	service := NewService(repository)

	got, err := service.GetByIDAndOwner(context.Background(), 42, 17)
	if err != nil {
		t.Fatalf("GetByIDAndOwner() error = %v, want nil", err)
	}
	if repository.getCalls != 1 || repository.queriedArenaID != 42 || repository.queriedOwnerID != 17 {
		t.Errorf("repository get calls = %d, arena ID = %d, owner ID = %d; want one call for arena 42 and owner 17", repository.getCalls, repository.queriedArenaID, repository.queriedOwnerID)
	}
	if got != want {
		t.Errorf("GetByIDAndOwner() = %+v, want %+v", got, want)
	}
}

func TestServiceGetByIDAndOwnerRejectsInvalidArenaID(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	_, err := service.GetByIDAndOwner(context.Background(), 0, 17)
	if !errors.Is(err, ErrInvalidArenaID) {
		t.Fatalf("GetByIDAndOwner() error = %v, want %v", err, ErrInvalidArenaID)
	}
	if repository.getCalls != 0 {
		t.Errorf("repository get calls = %d, want 0", repository.getCalls)
	}
}

func TestServiceGetByIDAndOwnerRejectsInvalidOwner(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	_, err := service.GetByIDAndOwner(context.Background(), 42, 0)
	if !errors.Is(err, ErrInvalidOwner) {
		t.Fatalf("GetByIDAndOwner() error = %v, want %v", err, ErrInvalidOwner)
	}
	if repository.getCalls != 0 {
		t.Errorf("repository get calls = %d, want 0", repository.getCalls)
	}
}

func TestServiceGetByIDAndOwnerReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	service := NewService(&fakeRepository{getError: wantErr})

	_, err := service.GetByIDAndOwner(context.Background(), 42, 17)
	if !errors.Is(err, wantErr) {
		t.Fatalf("GetByIDAndOwner() error = %v, want wrapped repository error %v", err, wantErr)
	}
}

func TestServiceUpdatePersistsFullArenaDetails(t *testing.T) {
	want := Arena{
		ID: 42, OwnerUserID: 17, Name: "Arena Central", ArenaType: "society",
		Address: "Rua do Esporte, 100", City: "São Paulo", State: "SP", Active: false,
	}
	repository := &fakeRepository{updatedArena: want}
	service := NewService(repository)
	details := Details{
		Name: " Arena Central ", ArenaType: " Society ", Address: " Rua do Esporte, 100 ",
		City: " São Paulo ", State: " sp ",
	}

	got, err := service.Update(context.Background(), 42, 17, details, false)
	if err != nil {
		t.Fatalf("Update() error = %v, want nil", err)
	}
	if repository.updateCalls != 1 || repository.updatedValue.ID != 42 || repository.updatedValue.OwnerUserID != 17 || repository.updatedValue.Active {
		t.Errorf("repository update calls = %d, value = %+v; want arena 42, owner 17 and inactive", repository.updateCalls, repository.updatedValue)
	}
	if got != want {
		t.Errorf("Update() = %+v, want %+v", got, want)
	}
}

func TestServiceUpdateRejectsInvalidArenaID(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	_, err := service.Update(context.Background(), 0, 17, validDetails(), true)
	if !errors.Is(err, ErrInvalidArenaID) {
		t.Fatalf("Update() error = %v, want %v", err, ErrInvalidArenaID)
	}
	if repository.updateCalls != 0 {
		t.Errorf("repository update calls = %d, want 0", repository.updateCalls)
	}
}

func TestServiceUpdateRejectsInvalidOwner(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	_, err := service.Update(context.Background(), 42, 0, validDetails(), true)
	if !errors.Is(err, ErrInvalidOwner) {
		t.Fatalf("Update() error = %v, want %v", err, ErrInvalidOwner)
	}
	if repository.updateCalls != 0 {
		t.Errorf("repository update calls = %d, want 0", repository.updateCalls)
	}
}

func TestServiceUpdateDoesNotPersistInvalidDetails(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)
	details := validDetails()
	details.Name = " "

	_, err := service.Update(context.Background(), 42, 17, details, true)
	if !errors.Is(err, ErrInvalidName) {
		t.Fatalf("Update() error = %v, want %v", err, ErrInvalidName)
	}
	if repository.updateCalls != 0 {
		t.Errorf("repository update calls = %d, want 0", repository.updateCalls)
	}
}

func TestServiceUpdateReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	service := NewService(&fakeRepository{updateError: wantErr})

	_, err := service.Update(context.Background(), 42, 17, validDetails(), true)
	if !errors.Is(err, wantErr) {
		t.Fatalf("Update() error = %v, want wrapped repository error %v", err, wantErr)
	}
}

func TestServiceToggleActiveReturnsRepositoryArena(t *testing.T) {
	want := Arena{ID: 42, OwnerUserID: 17, Active: false}
	repository := &fakeRepository{toggledArena: want}
	service := NewService(repository)

	got, err := service.ToggleActive(context.Background(), 42, 17)
	if err != nil {
		t.Fatalf("ToggleActive() error = %v, want nil", err)
	}
	if repository.toggleCalls != 1 || repository.toggleArenaID != 42 || repository.toggleOwnerID != 17 {
		t.Errorf("repository toggle calls = %d, arena ID = %d, owner ID = %d; want one call for arena 42 and owner 17", repository.toggleCalls, repository.toggleArenaID, repository.toggleOwnerID)
	}
	if got != want {
		t.Errorf("ToggleActive() = %+v, want %+v", got, want)
	}
}

func TestServiceToggleActiveRejectsInvalidArenaID(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	_, err := service.ToggleActive(context.Background(), 0, 17)
	if !errors.Is(err, ErrInvalidArenaID) {
		t.Fatalf("ToggleActive() error = %v, want %v", err, ErrInvalidArenaID)
	}
	if repository.toggleCalls != 0 {
		t.Errorf("repository toggle calls = %d, want 0", repository.toggleCalls)
	}
}

func TestServiceToggleActiveRejectsInvalidOwner(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository)

	_, err := service.ToggleActive(context.Background(), 42, 0)
	if !errors.Is(err, ErrInvalidOwner) {
		t.Fatalf("ToggleActive() error = %v, want %v", err, ErrInvalidOwner)
	}
	if repository.toggleCalls != 0 {
		t.Errorf("repository toggle calls = %d, want 0", repository.toggleCalls)
	}
}

func TestServiceToggleActiveReturnsRepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	service := NewService(&fakeRepository{toggleError: wantErr})

	_, err := service.ToggleActive(context.Background(), 42, 17)
	if !errors.Is(err, wantErr) {
		t.Fatalf("ToggleActive() error = %v, want wrapped repository error %v", err, wantErr)
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
	createdArena      Arena
	savedArena        Arena
	createCalls       int
	createError       error
	listedArenas      []Arena
	listedOwnerID     int64
	listCalls         int
	listError         error
	discoveredArenas  []PublicArena
	discoverFilters   DiscoveryFilters
	discoverCalls     int
	discoverError     error
	queriedArena      Arena
	queriedArenaID    int64
	queriedOwnerID    int64
	getCalls          int
	getError          error
	updatedValue      Arena
	updatedArena      Arena
	updateCalls       int
	updateError       error
	photoUpdateCalls  int
	photoUpdateURL    string
	updatedPhotoArena Arena
	photoUpdateError  error
	toggleArenaID     int64
	toggleOwnerID     int64
	toggledArena      Arena
	toggleCalls       int
	toggleError       error
}

func (r *fakeRepository) Update(_ context.Context, value Arena) (Arena, error) {
	r.updateCalls++
	r.updatedValue = value
	if r.updateError != nil {
		return Arena{}, r.updateError
	}
	return r.updatedArena, nil
}

func (r *fakeRepository) UpdatePhotoURL(_ context.Context, arenaID, ownerUserID int64, photoURL string) (Arena, error) {
	r.photoUpdateCalls++
	r.photoUpdateURL = photoURL
	if r.photoUpdateError != nil {
		return Arena{}, r.photoUpdateError
	}
	if r.updatedPhotoArena.ID == 0 {
		r.updatedPhotoArena = Arena{ID: arenaID, OwnerUserID: ownerUserID}
	}
	r.updatedPhotoArena.PhotoURL = &photoURL
	return r.updatedPhotoArena, nil
}

func (r *fakeRepository) ToggleActive(_ context.Context, arenaID, ownerUserID int64) (Arena, error) {
	r.toggleCalls++
	r.toggleArenaID = arenaID
	r.toggleOwnerID = ownerUserID
	if r.toggleError != nil {
		return Arena{}, r.toggleError
	}
	return r.toggledArena, nil
}

func (r *fakeRepository) GetByIDAndOwner(_ context.Context, arenaID, ownerUserID int64) (Arena, error) {
	r.getCalls++
	r.queriedArenaID = arenaID
	r.queriedOwnerID = ownerUserID
	if r.getError != nil {
		return Arena{}, r.getError
	}
	return r.queriedArena, nil
}

func (r *fakeRepository) ListByOwner(_ context.Context, ownerUserID int64) ([]Arena, error) {
	r.listCalls++
	r.listedOwnerID = ownerUserID
	if r.listError != nil {
		return nil, r.listError
	}
	return r.listedArenas, nil
}

func (r *fakeRepository) Discover(_ context.Context, filters DiscoveryFilters) ([]PublicArena, error) {
	r.discoverCalls++
	r.discoverFilters = filters
	if r.discoverError != nil {
		return nil, r.discoverError
	}
	return r.discoveredArenas, nil
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
	created.Active = value.Active
	return created, nil
}
