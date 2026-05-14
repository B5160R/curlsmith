package app

import (
	"errors"
	"testing"

	"github.com/b5160r/curlsmith/internal/domain"
)

type fakeCollectionLoader struct {
	loadedPath string
	collection domain.Collection
	err        error
}

func (f *fakeCollectionLoader) Load(path string) (domain.Collection, error) {
	f.loadedPath = path
	return f.collection, f.err
}

func (f *fakeCollectionLoader) Save(string, domain.Collection) error {
	return nil
}

func TestLoadCollectionUseCaseRequiresLoader(t *testing.T) {
	uc := LoadCollectionUseCase{}
	_, err := uc.Execute("./forge.json")
	if err == nil {
		t.Fatal("expected error when loader is nil")
	}
}

func TestLoadCollectionUseCaseRequiresPath(t *testing.T) {
	uc := LoadCollectionUseCase{Loader: &fakeCollectionLoader{}}
	_, err := uc.Execute("   ")
	if err == nil {
		t.Fatal("expected error when path is empty/whitespace")
	}
}

func TestLoadCollectionUseCaseReturnsCollection(t *testing.T) {
	want := domain.Collection{
		Variables: map[string]string{"base": "http://localhost"},
		Requests:  []domain.Request{{Name: "test", URL: "http://localhost/api"}},
	}
	loader := &fakeCollectionLoader{collection: want}
	uc := LoadCollectionUseCase{Loader: loader}

	got, err := uc.Execute("  ./forge.json  ")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if loader.loadedPath != "./forge.json" {
		t.Fatalf("expected trimmed path ./forge.json, got %q", loader.loadedPath)
	}

	if len(got.Requests) != 1 || got.Requests[0].Name != "test" {
		t.Fatalf("unexpected collection: %#v", got)
	}
}

func TestLoadCollectionUseCaseReturnsLoaderError(t *testing.T) {
	wantErr := errors.New("file corrupt")
	loader := &fakeCollectionLoader{err: wantErr}
	uc := LoadCollectionUseCase{Loader: loader}

	_, err := uc.Execute("./forge.json")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
}
