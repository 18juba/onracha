package arena

import (
	"context"
	"errors"
	"io"
)

const MaxPhotoSize = 5 << 20

var (
	ErrInvalidPhoto            = errors.New("arena photo is not a supported image")
	ErrPhotoTooLarge           = errors.New("arena photo exceeds the maximum size")
	ErrPhotoStorageUnavailable = errors.New("arena photo storage is unavailable")
	ErrPhotoCleanup            = errors.New("arena photo saved but previous photo cleanup failed")
)

// PhotoStorage stores validated arena photos outside the database.
type PhotoStorage interface {
	Store(ctx context.Context, source io.Reader) (string, error)
	Delete(ctx context.Context, photoURL string) error
}
