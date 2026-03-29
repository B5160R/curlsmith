package storage

import (
	"encoding/json"
	"errors"
	"os"
	"regexp"

	"github.com/b5160r/curlsmith/internal/domain"
)

type FileStore struct {
	Path string
}

type CollectionStore struct{}

func (CollectionStore) Load(path string) (domain.Collection, error) {
	return (&FileStore{Path: path}).Load()
}

func (CollectionStore) Save(path string, collection domain.Collection) error {
	return (&FileStore{Path: path}).Save(collection)
}

type CollectionLoader struct{}

func (CollectionLoader) Load(path string) (domain.Collection, error) {
	return (&FileStore{Path: path}).Load()
}

func (fs *FileStore) Save(collection domain.Collection) error {
	if fs.Path == "" {
		return errors.New("storage path is required")
	}

	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(fs.Path, data, 0o644)
}

func (fs *FileStore) Load() (domain.Collection, error) {
	if fs.Path == "" {
		return domain.Collection{}, errors.New("storage path is required")
	}

	data, err := os.ReadFile(fs.Path)
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

// {{base_url}} → actual value

var variablePattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_.-]+)\s*\}\}`)

func Resolve(input string, vars map[string]string) string {
	if len(vars) == 0 || input == "" {
		return input
	}

	return variablePattern.ReplaceAllStringFunc(input, func(match string) string {
		parts := variablePattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}

		value, ok := vars[parts[1]]
		if !ok {
			return match
		}

		return value
	})
}
