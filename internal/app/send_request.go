package app

import (
	"errors"
	"net/url"
	"strings"

	"github.com/b5160r/curlsmith/internal/domain"
)

type RequestSender interface {
	Do(req domain.Request) (domain.Response, error)
}

type SendRequestUseCase struct {
	Client RequestSender
}

func (uc *SendRequestUseCase) Execute(req domain.Request) (domain.Response, error) {
	if uc.Client == nil {
		return domain.Response{}, errors.New("request sender is required")
	}

	req.Method = strings.ToUpper(strings.TrimSpace(req.Method))
	if req.Method == "" {
		req.Method = "GET"
	}

	req.URL = strings.TrimSpace(req.URL)
	if req.URL == "" {
		return domain.Response{}, errors.New("request URL is required")
	}

	if _, err := url.ParseRequestURI(req.URL); err != nil {
		return domain.Response{}, err
	}

	if req.Headers == nil {
		req.Headers = map[string][]string{}
	}

	return uc.Client.Do(req)
}
