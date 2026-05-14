package http

import (
	"bytes"
	"context"
	"io"
	stdhttp "net/http"
	"time"

	"github.com/b5160r/curlsmith/internal/domain"
)

const DefaultTimeout = 30 * time.Second

type Client struct {
	HTTPClient *stdhttp.Client
	Timeout    time.Duration
}

func (c *Client) Do(req domain.Request) (domain.Response, error) {
	body := bytes.NewBufferString(req.Body)

	timeout := c.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	httpReq, err := stdhttp.NewRequestWithContext(ctx, req.Method, req.URL, body)
	if err != nil {
		return domain.Response{}, err
	}

	for key, values := range req.Headers {
		for _, value := range values {
			httpReq.Header.Add(key, value)
		}
	}

	client := c.HTTPClient
	if client == nil {
		client = stdhttp.DefaultClient
	}

	start := time.Now()
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return domain.Response{}, err
	}
	defer httpResp.Body.Close()
	duration := time.Since(start)

	responseBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return domain.Response{}, err
	}

	return domain.Response{
		Status:   httpResp.StatusCode,
		Headers:  map[string][]string(httpResp.Header.Clone()),
		Body:     responseBody,
		Duration: duration,
	}, nil
}
