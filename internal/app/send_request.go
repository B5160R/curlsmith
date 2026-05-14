package app

import (
	"errors"
	"net/url"
	"strings"

	"github.com/b5160r/curlsmith/internal/domain"
)

type SendRequestUseCase struct {
	Client   RequestSender
	Vars     map[string]string
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

	// Resolve collection variables in URL and body.
	if len(uc.Vars) > 0 {
		req.URL = domain.Resolve(req.URL, uc.Vars)
		req.Body = domain.Resolve(req.Body, uc.Vars)
		for key, values := range req.Headers {
			for i, v := range values {
				req.Headers[key][i] = domain.Resolve(v, uc.Vars)
			}
		}
	}

	if _, err := url.ParseRequestURI(req.URL); err != nil {
		return domain.Response{}, err
	}

	if req.Headers == nil {
		req.Headers = map[string][]string{}
	}

	return uc.Client.Do(req)
}
