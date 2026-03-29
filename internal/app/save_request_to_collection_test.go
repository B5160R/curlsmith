package app

import (
	"errors"
	"reflect"
	"testing"

	"github.com/b5160r/curlsmith/internal/domain"
)

type fakeCollectionStore struct {
	loadedPath    string
	savedPath     string
	loadedValue   domain.Collection
	savedValue    domain.Collection
	loadErr       error
	saveErr       error
	saveWasCalled bool
}

func (f *fakeCollectionStore) Load(path string) (domain.Collection, error) {
	f.loadedPath = path
	return f.loadedValue, f.loadErr
}

func (f *fakeCollectionStore) Save(path string, collection domain.Collection) error {
	f.savedPath = path
	f.savedValue = collection
	f.saveWasCalled = true
	return f.saveErr
}

func TestSaveRequestToCollectionUseCaseCreatesCollectionEntry(t *testing.T) {
	store := &fakeCollectionStore{}
	useCase := SaveRequestToCollectionUseCase{Store: store}

	collection, err := useCase.Execute("./forge.json", domain.Request{
		Name:   "list swords",
		Method: "get",
		URL:    " https://example.com/swords ",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if store.loadedPath != "./forge.json" {
		t.Fatalf("expected load path ./forge.json, got %q", store.loadedPath)
	}

	if !store.saveWasCalled {
		t.Fatal("expected Save to be called")
	}

	if len(collection.Requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(collection.Requests))
	}

	saved := collection.Requests[0]
	if saved.Method != "GET" {
		t.Fatalf("expected normalized method GET, got %q", saved.Method)
	}

	if saved.URL != "https://example.com/swords" {
		t.Fatalf("expected trimmed URL, got %q", saved.URL)
	}

	if !reflect.DeepEqual(saved.Headers, map[string][]string{}) {
		t.Fatalf("expected headers to be initialized, got %#v", saved.Headers)
	}
}

func TestSaveRequestToCollectionUseCaseUpdatesMatchingRequestByName(t *testing.T) {
	store := &fakeCollectionStore{
		loadedValue: domain.Collection{
			Requests: []domain.Request{{
				Name:   "list swords",
				Method: "GET",
				URL:    "https://example.com/old",
			}},
		},
	}
	useCase := SaveRequestToCollectionUseCase{Store: store}

	collection, err := useCase.Execute("./forge.json", domain.Request{
		Name:   "list swords",
		Method: "POST",
		URL:    "https://example.com/new",
		Body:   "tempered",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if len(collection.Requests) != 1 {
		t.Fatalf("expected 1 request after update, got %d", len(collection.Requests))
	}

	updated := collection.Requests[0]
	if updated.URL != "https://example.com/new" {
		t.Fatalf("expected updated URL, got %q", updated.URL)
	}

	if updated.Method != "POST" {
		t.Fatalf("expected updated method POST, got %q", updated.Method)
	}
}

func TestSaveRequestToCollectionUseCaseReturnsStoreError(t *testing.T) {
	wantErr := errors.New("disk full")
	store := &fakeCollectionStore{saveErr: wantErr}
	useCase := SaveRequestToCollectionUseCase{Store: store}

	_, err := useCase.Execute("./forge.json", domain.Request{URL: "https://example.com"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
}
