package storage

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/b5160r/curlsmith/internal/domain"
)

func TestFileStoreSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "collection.json")
	store := &FileStore{Path: path}

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

	if err := store.Save(want); err != nil {
		t.Fatalf("Save returned error: %v", err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("loaded collection mismatch\nwant: %#v\ngot: %#v", want, got)
	}
}

func TestFileStoreLoadMissingFileReturnsEmptyCollection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	store := &FileStore{Path: path}

	got, err := store.Load()
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

func TestResolveReplacesKnownVariables(t *testing.T) {
	input := "{{base_url}}/users/{{user_id}}"
	vars := map[string]string{
		"base_url": "https://api.example.com",
		"user_id":  "42",
	}

	got := Resolve(input, vars)
	want := "https://api.example.com/users/42"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestResolveLeavesUnknownVariables(t *testing.T) {
	got := Resolve("{{base_url}}/{{missing}}", map[string]string{"base_url": "https://api.example.com"})
	want := "https://api.example.com/{{missing}}"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
