package photostorage

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"onracha/internal/arena"
)

const (
	MaxPhotoDimension = 4096
	maxPhotoPixels    = 16 * 1024 * 1024
)

// LocalStorage stores validated photos in a directory served by the API.
type LocalStorage struct {
	directory     string
	publicBaseURL string
	publicHost    string
	publicScheme  string
}

// Directory returns the absolute path used for uploaded files.
func (s *LocalStorage) Directory() string {
	return s.directory
}

// NewLocalStorage creates the upload directory and configures the public API URL used in photo references.
func NewLocalStorage(directory, publicBaseURL string) (*LocalStorage, error) {
	if strings.TrimSpace(directory) == "" {
		return nil, errors.New("photo storage directory is required")
	}

	parsedBaseURL, err := url.Parse(publicBaseURL)
	if err != nil || parsedBaseURL.Hostname() == "" ||
		(!strings.EqualFold(parsedBaseURL.Scheme, "http") && !strings.EqualFold(parsedBaseURL.Scheme, "https")) ||
		(parsedBaseURL.Path != "" && parsedBaseURL.Path != "/") || parsedBaseURL.RawQuery != "" || parsedBaseURL.Fragment != "" {
		return nil, errors.New("photo public URL must be an HTTP or HTTPS origin")
	}

	absoluteDirectory, err := filepath.Abs(directory)
	if err != nil {
		return nil, fmt.Errorf("resolve photo storage directory: %w", err)
	}
	if err := os.MkdirAll(absoluteDirectory, 0o750); err != nil {
		return nil, fmt.Errorf("create photo storage directory: %w", err)
	}

	return &LocalStorage{
		directory:     absoluteDirectory,
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
		publicHost:    parsedBaseURL.Host,
		publicScheme:  parsedBaseURL.Scheme,
	}, nil
}

// Store validates the image bytes and writes the image under a generated filename.
func (s *LocalStorage) Store(ctx context.Context, source io.Reader) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if source == nil {
		return "", arena.ErrInvalidPhoto
	}

	data, err := io.ReadAll(io.LimitReader(source, arena.MaxPhotoSize+1))
	if err != nil {
		return "", fmt.Errorf("read arena photo: %w", err)
	}
	if len(data) > arena.MaxPhotoSize {
		return "", arena.ErrPhotoTooLarge
	}
	if len(data) == 0 {
		return "", arena.ErrInvalidPhoto
	}

	configuration, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || !supportedImageFormat(format) || !validImageDimensions(configuration) {
		return "", arena.ErrInvalidPhoto
	}
	_, decodedFormat, err := image.Decode(bytes.NewReader(data))
	if err != nil || decodedFormat != format {
		return "", arena.ErrInvalidPhoto
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	identifier := make([]byte, 16)
	if _, err := rand.Read(identifier); err != nil {
		return "", fmt.Errorf("generate arena photo name: %w", err)
	}
	filename := hex.EncodeToString(identifier) + imageExtension(format)

	temporaryFile, err := os.CreateTemp(s.directory, ".arena-photo-*")
	if err != nil {
		return "", fmt.Errorf("create temporary arena photo: %w", err)
	}
	temporaryPath := temporaryFile.Name()
	defer os.Remove(temporaryPath)
	if _, err := temporaryFile.Write(data); err != nil {
		_ = temporaryFile.Close()
		return "", fmt.Errorf("write temporary arena photo: %w", err)
	}
	if err := temporaryFile.Close(); err != nil {
		return "", fmt.Errorf("close temporary arena photo: %w", err)
	}
	if err := os.Rename(temporaryPath, filepath.Join(s.directory, filename)); err != nil {
		return "", fmt.Errorf("save arena photo: %w", err)
	}

	return s.publicBaseURL + "/uploads/" + filename, nil
}

// Delete removes a photo only when its reference belongs to this local storage.
func (s *LocalStorage) Delete(ctx context.Context, photoURL string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	parsedURL, err := url.Parse(photoURL)
	if err != nil || !strings.EqualFold(parsedURL.Scheme, s.publicScheme) || !strings.EqualFold(parsedURL.Host, s.publicHost) ||
		parsedURL.RawQuery != "" || parsedURL.Fragment != "" || !strings.HasPrefix(parsedURL.Path, "/uploads/") {
		return nil
	}

	filename := strings.TrimPrefix(parsedURL.Path, "/uploads/")
	if !isGeneratedPhotoFilename(filename) {
		return nil
	}
	if err := os.Remove(filepath.Join(s.directory, filename)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete arena photo: %w", err)
	}
	return nil
}

func supportedImageFormat(format string) bool {
	return format == "jpeg" || format == "png"
}

func validImageDimensions(configuration image.Config) bool {
	if configuration.Width <= 0 || configuration.Height <= 0 ||
		configuration.Width > MaxPhotoDimension || configuration.Height > MaxPhotoDimension {
		return false
	}
	return int64(configuration.Width)*int64(configuration.Height) <= maxPhotoPixels
}

func imageExtension(format string) string {
	if format == "jpeg" {
		return ".jpg"
	}
	return ".png"
}

func isGeneratedPhotoFilename(filename string) bool {
	extension := filepath.Ext(filename)
	if extension != ".jpg" && extension != ".png" {
		return false
	}
	identifier := strings.TrimSuffix(filename, extension)
	if len(identifier) != 32 {
		return false
	}
	_, err := hex.DecodeString(identifier)
	return err == nil
}
