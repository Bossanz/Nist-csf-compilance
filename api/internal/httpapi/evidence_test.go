package httpapi

import (
	"io"
	"strings"
	"testing"
)

func TestLocalEvidenceStorageWritesAndRemovesOpaqueFile(t *testing.T) {
	storage := newLocalEvidenceStorage(t.TempDir())
	key, size, err := storage.Save(strings.NewReader("evidence"))
	if err != nil || size != 8 || key == "" || strings.Contains(key, "evidence") {
		t.Fatalf("save evidence: key=%s size=%d err=%v", key, size, err)
	}
	file, err := storage.Open(key)
	if err != nil {
		t.Fatal(err)
	}
	contents, err := io.ReadAll(file)
	file.Close()
	if err != nil || string(contents) != "evidence" {
		t.Fatalf("read evidence: %q err=%v", contents, err)
	}
	if err := storage.Remove(key); err != nil {
		t.Fatal(err)
	}
	if err := storage.Remove(key); err != nil {
		t.Fatalf("remove should be idempotent: %v", err)
	}
}

func TestLocalEvidenceStorageRejectsOversizeFile(t *testing.T) {
	storage := newLocalEvidenceStorage(t.TempDir())
	if _, _, err := storage.Save(io.LimitReader(endlessReader{}, maxEvidenceBytes+1)); err != ErrEvidenceTooLarge {
		t.Fatalf("expected oversize error, got %v", err)
	}
}

type endlessReader struct{}

func (endlessReader) Read(buffer []byte) (int, error) {
	for index := range buffer {
		buffer[index] = 'x'
	}
	return len(buffer), nil
}

func TestLocalEvidenceStorageRejectsPathTraversal(t *testing.T) {
	storage := newLocalEvidenceStorage(t.TempDir())
	if _, err := storage.Open("../outside"); err != ErrInvalidEvidenceStorage {
		t.Fatalf("expected invalid storage key, got %v", err)
	}
}
