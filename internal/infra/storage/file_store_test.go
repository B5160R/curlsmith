package storage

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/b5160r/curlsmith/internal/domain"
)

func TestFileStoreSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "collection.json")
	store := FileStore{}

	want := domain.Collection{
		Variables: map[string]string{"base_url": "https://api.example.com"},
		Requests: []domain.Request{
			{
				ID:     "1",
				Name:   "list items",
				Method: "GET",
				URL:    "{{base_url}}/items",
			},
		},
	}

	if err := store.Save(path, want); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	got, err := store.Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("loaded collection mismatch\nwant: %#v\ngot: %#v", want, got)
	}
}

func TestFileStoreLoadMissingFileReturnsEmptyCollection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	store := FileStore{}

	got, err := store.Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if got.Variables != nil {
		t.Fatalf("expected nil variables for missing file, got %#v", got.Variables)
	}

	if got.Requests != nil {
		t.Fatalf("expected nil requests for missing file, got %#v", got.Requests)
	}
}

func TestFileStoreSaveEmptyPathReturnsError(t *testing.T) {
	store := FileStore{}
	err := store.Save("", domain.Collection{})
	if err == nil {
		t.Fatal("expected error when path is empty")
	}
}

func TestFileStoreLoadEmptyPathReturnsError(t *testing.T) {
	store := FileStore{}
	_, err := store.Load("")
	if err == nil {
		t.Fatal("expected error when path is empty")
	}
}

func TestFileStoreLoadInvalidJSONReturnsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{not valid json"), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	store := FileStore{}
	_, err := store.Load(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
