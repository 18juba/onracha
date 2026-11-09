package photostorage

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"onracha/internal/arena"
)

func TestLocalStorageStoresValidJPEGAndPNG(t *testing.T) {
	for _, format := range []string{"jpeg", "png"} {
		t.Run(format, func(t *testing.T) {
			storage, directory := newTestLocalStorage(t)
			data := encodeTestImage(t, format, image.Rect(0, 0, 2, 2))

			photoURL, err := storage.Store(context.Background(), bytes.NewReader(data))
			if err != nil {
				t.Fatalf("Store() error = %v, want nil", err)
			}

			parsedURL, err := url.Parse(photoURL)
			if err != nil {
				t.Fatalf("parse returned photo URL: %v", err)
			}
			if parsedURL.Host != "api.example.test" || !strings.HasPrefix(parsedURL.Path, "/uploads/") {
				t.Errorf("photo URL = %q, want public upload URL", photoURL)
			}
			filePath := filepath.Join(directory, filepath.Base(parsedURL.Path))
			stored, err := os.ReadFile(filePath)
			if err != nil {
				t.Fatalf("read stored photo: %v", err)
			}
			_, actualFormat, err := image.Decode(bytes.NewReader(stored))
			if err != nil || actualFormat != format {
				t.Errorf("stored image format = %q, error = %v; want %q", actualFormat, err, format)
			}
		})
	}
}

func TestLocalStorageRejectsInvalidAndUnsupportedImageContent(t *testing.T) {
	storage, _ := newTestLocalStorage(t)
	validPNG := encodeTestImage(t, "png", image.Rect(0, 0, 2, 2))
	truncatedPNG := validPNG[:len(validPNG)-8]
	validGIF := encodeTestImage(t, "gif", image.Rect(0, 0, 2, 2))

	for _, test := range []struct {
		name string
		data []byte
	}{
		{name: "plain text", data: []byte("not an image")},
		{name: "truncated PNG", data: truncatedPNG},
		{name: "unsupported GIF", data: validGIF},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := storage.Store(context.Background(), bytes.NewReader(test.data)); !errors.Is(err, arena.ErrInvalidPhoto) {
				t.Errorf("Store() error = %v, want %v", err, arena.ErrInvalidPhoto)
			}
		})
	}
}

func TestLocalStorageRejectsFilesOverTheSizeLimit(t *testing.T) {
	storage, _ := newTestLocalStorage(t)
	data := bytes.Repeat([]byte("x"), arena.MaxPhotoSize+1)

	if _, err := storage.Store(context.Background(), bytes.NewReader(data)); !errors.Is(err, arena.ErrPhotoTooLarge) {
		t.Errorf("Store() error = %v, want %v", err, arena.ErrPhotoTooLarge)
	}
}

func TestLocalStorageRejectsImagesWithExcessiveDimensions(t *testing.T) {
	storage, _ := newTestLocalStorage(t)
	data := encodeTestImage(t, "png", image.Rect(0, 0, MaxPhotoDimension+1, 1))

	if _, err := storage.Store(context.Background(), bytes.NewReader(data)); !errors.Is(err, arena.ErrInvalidPhoto) {
		t.Errorf("Store() error = %v, want %v", err, arena.ErrInvalidPhoto)
	}
}

func TestLocalStorageDeleteRemovesOnlyFilesInItsUploadDirectory(t *testing.T) {
	storage, directory := newTestLocalStorage(t)
	data := encodeTestImage(t, "png", image.Rect(0, 0, 2, 2))
	photoURL, err := storage.Store(context.Background(), bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Store() error = %v, want nil", err)
	}

	parsedURL, _ := url.Parse(photoURL)
	filePath := filepath.Join(directory, filepath.Base(parsedURL.Path))
	if err := storage.Delete(context.Background(), photoURL); err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("stored photo still exists after Delete(): stat error = %v", err)
	}
	if err := storage.Delete(context.Background(), "https://external.example.test/photo.png"); err != nil {
		t.Errorf("Delete() external URL error = %v, want nil", err)
	}
}

func newTestLocalStorage(t *testing.T) (*LocalStorage, string) {
	t.Helper()
	directory := t.TempDir()
	storage, err := NewLocalStorage(directory, "http://api.example.test")
	if err != nil {
		t.Fatalf("NewLocalStorage() error = %v, want nil", err)
	}
	return storage, directory
}

func encodeTestImage(t *testing.T, format string, bounds image.Rectangle) []byte {
	t.Helper()
	imageValue := image.NewRGBA(bounds)
	var buffer bytes.Buffer
	var err error
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buffer, imageValue, &jpeg.Options{Quality: 80})
	case "png":
		err = png.Encode(&buffer, imageValue)
	case "gif":
		err = gif.Encode(&buffer, imageValue, nil)
	default:
		t.Fatalf("unsupported test image format %q", format)
	}
	if err != nil {
		t.Fatalf("encode %s test image: %v", format, err)
	}
	return buffer.Bytes()
}
