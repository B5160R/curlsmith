package app

import "github.com/b5160r/curlsmith/internal/domain"

// RequestSender is the port for sending HTTP requests.
type RequestSender interface {
	Do(req domain.Request) (domain.Response, error)
}

// CollectionStore is the port for loading and saving collections.
type CollectionStore interface {
	Load(path string) (domain.Collection, error)
	Save(path string, collection domain.Collection) error
}
