package storage

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/b5160r/curlsmith/internal/domain"
)

// FileStore implements app.CollectionStore backed by JSON files.
type FileStore struct{}

func (FileStore) Save(path string, collection domain.Collection) error {
	if path == "" {
		return errors.New("storage path is required")
	}

	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func (FileStore) Load(path string) (domain.Collection, error) {
	if path == "" {
		return domain.Collection{}, errors.New("storage path is required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return domain.Collection{}, nil
		}

		return domain.Collection{}, err
	}

	var collection domain.Collection
	if err := json.Unmarshal(data, &collection); err != nil {
		return domain.Collection{}, err
	}

	if collection.Variables == nil {
		collection.Variables = map[string]string{}
	}

	if collection.Requests == nil {
		collection.Requests = []domain.Request{}
	}

	return collection, nil
}
