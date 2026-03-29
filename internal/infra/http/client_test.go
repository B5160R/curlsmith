package http

import (
	"errors"
	"io"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/b5160r/curlsmith/internal/domain"
)

type roundTripFunc func(*stdhttp.Request) (*stdhttp.Response, error)

func (fn roundTripFunc) RoundTrip(req *stdhttp.Request) (*stdhttp.Response, error) {
	return fn(req)
}

func TestClientDoMapsRequestAndResponse(t *testing.T) {
	server := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		if r.Method != stdhttp.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		if got := r.Header.Get("X-Test"); got != "1" {
			t.Fatalf("expected X-Test header to be 1, got %q", got)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed reading request body: %v", err)
		}

		if string(body) != "hello" {
			t.Fatalf("expected body hello, got %q", string(body))
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(stdhttp.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	client := &Client{HTTPClient: server.Client()}
	response, err := client.Do(domain.Request{
		Method:  "POST",
		URL:     server.URL,
		Headers: map[string][]string{"X-Test": {"1"}},
		Body:    "hello",
	})
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}

	if response.Status != stdhttp.StatusCreated {
		t.Fatalf("expected status %d, got %d", stdhttp.StatusCreated, response.Status)
	}

	if got := string(response.Body); got != "ok" {
		t.Fatalf("expected response body ok, got %q", got)
	}

	if got := response.Headers["Content-Type"]; len(got) != 1 || got[0] != "text/plain" {
		t.Fatalf("expected Content-Type header, got %#v", got)
	}
}

func TestClientDoReturnsHTTPClientError(t *testing.T) {
	wantErr := errors.New("network down")
	client := &Client{
		HTTPClient: &stdhttp.Client{
			Transport: roundTripFunc(func(req *stdhttp.Request) (*stdhttp.Response, error) {
				return nil, wantErr
			}),
		},
	}

	_, err := client.Do(domain.Request{
		Method: "GET",
		URL:    "https://example.com",
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
}
