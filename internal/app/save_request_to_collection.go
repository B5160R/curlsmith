package app

import (
	"errors"
	"strings"

	"github.com/b5160r/curlsmith/internal/domain"
)

type CollectionStore interface {
	Load(path string) (domain.Collection, error)
	Save(path string, collection domain.Collection) error
}

type SaveRequestToCollectionUseCase struct {
	Store CollectionStore
}

func (uc *SaveRequestToCollectionUseCase) Execute(path string, request domain.Request) (domain.Collection, error) {
	if uc.Store == nil {
		return domain.Collection{}, errors.New("collection store is required")
	}

	path = strings.TrimSpace(path)
	if path == "" {
		return domain.Collection{}, errors.New("collection path is required")
	}

	request = normalizeRequestForCollection(request)
	if request.URL == "" {
		return domain.Collection{}, errors.New("request URL is required")
	}

	collection, err := uc.Store.Load(path)
	if err != nil {
		return domain.Collection{}, err
	}

	if collection.Variables == nil {
		collection.Variables = map[string]string{}
	}
	if collection.Requests == nil {
		collection.Requests = []domain.Request{}
	}

	index := findRequestIndex(collection.Requests, request)
	if index >= 0 {
		collection.Requests[index] = request
	} else {
		collection.Requests = append(collection.Requests, request)
	}

	if err := uc.Store.Save(path, collection); err != nil {
		return domain.Collection{}, err
	}

	return collection, nil
}

func normalizeRequestForCollection(request domain.Request) domain.Request {
	request.ID = strings.TrimSpace(request.ID)
	request.Name = strings.TrimSpace(request.Name)
	request.Method = strings.ToUpper(strings.TrimSpace(request.Method))
	request.URL = strings.TrimSpace(request.URL)
	if request.Headers == nil {
		request.Headers = map[string][]string{}
	}
	return request
}

func findRequestIndex(requests []domain.Request, target domain.Request) int {
	if target.ID != "" {
		for index, request := range requests {
			if strings.TrimSpace(request.ID) == target.ID {
				return index
			}
		}
	}

	if target.Name != "" {
		for index, request := range requests {
			if strings.EqualFold(strings.TrimSpace(request.Name), target.Name) {
				return index
			}
		}
	}

	return -1
}
