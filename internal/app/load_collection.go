package app

import (
	"errors"
	"strings"

	"github.com/b5160r/curlsmith/internal/domain"
)

type LoadCollectionUseCase struct {
	Loader CollectionStore
}

func (uc *LoadCollectionUseCase) Execute(path string) (domain.Collection, error) {
	if uc.Loader == nil {
		return domain.Collection{}, errors.New("collection loader is required")
	}

	path = strings.TrimSpace(path)
	if path == "" {
		return domain.Collection{}, errors.New("collection path is required")
	}

	return uc.Loader.Load(path)
}
