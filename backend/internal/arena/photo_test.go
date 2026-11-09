package arena

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestServiceReplacePhotoStoresReferenceAndRemovesPreviousPhoto(t *testing.T) {
	previousURL := "http://localhost:8080/uploads/previous.png"
	newURL := "http://localhost:8080/uploads/new.png"
	repository := &fakeRepository{
		queriedArena:      Arena{ID: 42, OwnerUserID: 17, PhotoURL: &previousURL},
		updatedPhotoArena: Arena{ID: 42, OwnerUserID: 17, PhotoURL: &newURL},
	}
	storage := &fakePhotoStorage{storedURL: newURL}
	service := NewService(repository, storage)

	updated, err := service.ReplacePhoto(context.Background(), 42, 17, strings.NewReader("image bytes"))
	if err != nil {
		t.Fatalf("ReplacePhoto() error = %v, want nil", err)
	}
	if updated.PhotoURL == nil || *updated.PhotoURL != newURL {
		t.Errorf("ReplacePhoto().PhotoURL = %v, want %q", updated.PhotoURL, newURL)
	}
	if storage.storeCalls != 1 || repository.photoUpdateCalls != 1 || repository.photoUpdateURL != newURL {
		t.Errorf("storage calls = %d, repository photo updates = %d for %q; want one update to new URL", storage.storeCalls, repository.photoUpdateCalls, repository.photoUpdateURL)
	}
	if len(storage.deletedURLs) != 1 || storage.deletedURLs[0] != previousURL {
		t.Errorf("deleted photo URLs = %v, want previous URL %q", storage.deletedURLs, previousURL)
	}
}

func TestServiceReplacePhotoRejectsInvalidIDsBeforeStorage(t *testing.T) {
	storage := &fakePhotoStorage{storedURL: "http://localhost:8080/uploads/new.png"}
	service := NewService(&fakeRepository{}, storage)

	for _, test := range []struct {
		name        string
		arenaID     int64
		ownerUserID int64
		wantErr     error
	}{
		{name: "arena ID", arenaID: 0, ownerUserID: 17, wantErr: ErrInvalidArenaID},
		{name: "owner ID", arenaID: 42, ownerUserID: 0, wantErr: ErrInvalidOwner},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := service.ReplacePhoto(context.Background(), test.arenaID, test.ownerUserID, strings.NewReader("image")); !errors.Is(err, test.wantErr) {
				t.Errorf("ReplacePhoto() error = %v, want %v", err, test.wantErr)
			}
		})
	}
	if storage.storeCalls != 0 {
		t.Errorf("storage Store() calls = %d, want 0 for invalid IDs", storage.storeCalls)
	}
}

func TestServiceReplacePhotoReturnsStorageValidationError(t *testing.T) {
	wantErr := ErrInvalidPhoto
	storage := &fakePhotoStorage{storeError: wantErr}
	repository := &fakeRepository{queriedArena: Arena{ID: 42, OwnerUserID: 17}}
	service := NewService(repository, storage)

	if _, err := service.ReplacePhoto(context.Background(), 42, 17, strings.NewReader("bad image")); !errors.Is(err, wantErr) {
		t.Errorf("ReplacePhoto() error = %v, want %v", err, wantErr)
	}
	if repository.photoUpdateCalls != 0 {
		t.Errorf("repository photo updates = %d, want 0", repository.photoUpdateCalls)
	}
}

func TestServiceReplacePhotoRejectsInvalidStorageReference(t *testing.T) {
	storage := &fakePhotoStorage{storedURL: "javascript:alert(1)"}
	repository := &fakeRepository{queriedArena: Arena{ID: 42, OwnerUserID: 17}}
	service := NewService(repository, storage)

	if _, err := service.ReplacePhoto(context.Background(), 42, 17, strings.NewReader("image")); !errors.Is(err, ErrInvalidPhotoURL) {
		t.Errorf("ReplacePhoto() error = %v, want %v", err, ErrInvalidPhotoURL)
	}
	if repository.photoUpdateCalls != 0 || len(storage.deletedURLs) != 1 {
		t.Errorf("repository updates = %d, deleted URLs = %v; want no update and one cleanup", repository.photoUpdateCalls, storage.deletedURLs)
	}
}

func TestServiceReplacePhotoDeletesNewFileWhenReferenceUpdateFails(t *testing.T) {
	newURL := "http://localhost:8080/uploads/new.png"
	wantErr := errors.New("database unavailable")
	storage := &fakePhotoStorage{storedURL: newURL}
	repository := &fakeRepository{queriedArena: Arena{ID: 42, OwnerUserID: 17}, photoUpdateError: wantErr}
	service := NewService(repository, storage)

	if _, err := service.ReplacePhoto(context.Background(), 42, 17, strings.NewReader("image")); !errors.Is(err, wantErr) {
		t.Errorf("ReplacePhoto() error = %v, want %v", err, wantErr)
	}
	if len(storage.deletedURLs) != 1 || storage.deletedURLs[0] != newURL {
		t.Errorf("deleted photo URLs = %v, want failed upload %q", storage.deletedURLs, newURL)
	}
}

func TestServiceReplacePhotoReportsPreviousFileCleanupFailure(t *testing.T) {
	previousURL := "http://localhost:8080/uploads/previous.png"
	newURL := "http://localhost:8080/uploads/new.png"
	cleanupErr := errors.New("permission denied")
	storage := &fakePhotoStorage{storedURL: newURL, deleteError: cleanupErr}
	repository := &fakeRepository{
		queriedArena:      Arena{ID: 42, OwnerUserID: 17, PhotoURL: &previousURL},
		updatedPhotoArena: Arena{ID: 42, OwnerUserID: 17, PhotoURL: &newURL},
	}
	service := NewService(repository, storage)

	updated, err := service.ReplacePhoto(context.Background(), 42, 17, strings.NewReader("image"))
	if !errors.Is(err, ErrPhotoCleanup) || !errors.Is(err, cleanupErr) {
		t.Errorf("ReplacePhoto() error = %v, want cleanup error %v", err, cleanupErr)
	}
	if updated.PhotoURL == nil || *updated.PhotoURL != newURL {
		t.Errorf("ReplacePhoto() updated URL = %v, want persisted new URL %q", updated.PhotoURL, newURL)
	}
}

type fakePhotoStorage struct {
	storedURL   string
	storeCalls  int
	storeError  error
	deletedURLs []string
	deleteError error
}

func (s *fakePhotoStorage) Store(_ context.Context, source io.Reader) (string, error) {
	s.storeCalls++
	_, _ = io.Copy(io.Discard, source)
	if s.storeError != nil {
		return "", s.storeError
	}
	return s.storedURL, nil
}

func (s *fakePhotoStorage) Delete(_ context.Context, photoURL string) error {
	s.deletedURLs = append(s.deletedURLs, photoURL)
	return s.deleteError
}
