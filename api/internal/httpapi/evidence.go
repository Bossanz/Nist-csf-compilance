package httpapi

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const maxEvidenceBytes int64 = 20 * 1024 * 1024

var (
	ErrEvidenceTooLarge       = errors.New("evidence file exceeds 20 MB")
	ErrUnsupportedEvidence    = errors.New("unsupported evidence file type")
	ErrInvalidEvidenceStorage = errors.New("invalid evidence storage key")
)

type localEvidenceStorage struct {
	dir string
}

func newLocalEvidenceStorage(dir string) *localEvidenceStorage {
	if dir == "" {
		dir = os.Getenv("EVIDENCE_DIR")
	}
	if dir == "" {
		dir = "/data/evidence"
	}
	return &localEvidenceStorage{dir: dir}
}

func (s *localEvidenceStorage) Save(source io.Reader) (string, int64, error) {
	if err := os.MkdirAll(s.dir, 0o750); err != nil {
		return "", 0, err
	}
	temporary, err := os.CreateTemp(s.dir, ".upload-*")
	if err != nil {
		return "", 0, err
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)

	size, copyErr := io.Copy(temporary, io.LimitReader(source, maxEvidenceBytes+1))
	closeErr := temporary.Close()
	if copyErr != nil {
		return "", 0, copyErr
	}
	if closeErr != nil {
		return "", 0, closeErr
	}
	if size > maxEvidenceBytes {
		return "", 0, ErrEvidenceTooLarge
	}

	storageKey := uuid.NewString()
	if err := os.Rename(temporaryName, filepath.Join(s.dir, storageKey)); err != nil {
		return "", 0, err
	}
	return storageKey, size, nil
}

func (s *localEvidenceStorage) Open(storageKey string) (io.ReadCloser, error) {
	path, err := s.path(storageKey)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

func (s *localEvidenceStorage) Remove(storageKey string) error {
	path, err := s.path(storageKey)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *localEvidenceStorage) path(storageKey string) (string, error) {
	if storageKey == "" || filepath.Base(storageKey) != storageKey || strings.ContainsAny(storageKey, `/\`) {
		return "", ErrInvalidEvidenceStorage
	}
	return filepath.Join(s.dir, storageKey), nil
}
