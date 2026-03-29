package app

import (
	"errors"
	"reflect"
	"testing"

	"github.com/b5160r/curlsmith/internal/domain"
)

type fakeSender struct {
	gotRequest domain.Request
	response   domain.Response
	err        error
}

func (f *fakeSender) Do(req domain.Request) (domain.Response, error) {
	f.gotRequest = req
	return f.response, f.err
}

func TestSendRequestUseCaseExecuteRequiresClient(t *testing.T) {
	useCase := SendRequestUseCase{}

	_, err := useCase.Execute(domain.Request{URL: "https://example.com"})
	if err == nil {
		t.Fatal("expected an error when client is nil")
	}
}

func TestSendRequestUseCaseExecuteRequiresURL(t *testing.T) {
	useCase := SendRequestUseCase{Client: &fakeSender{}}

	_, err := useCase.Execute(domain.Request{Method: "GET"})
	if err == nil {
		t.Fatal("expected an error when URL is empty")
	}
}

func TestSendRequestUseCaseExecuteNormalizesAndDelegates(t *testing.T) {
	fake := &fakeSender{
		response: domain.Response{Status: 201},
	}
	useCase := SendRequestUseCase{Client: fake}

	response, err := useCase.Execute(domain.Request{
		Method: " post ",
		URL:    "  https://example.com/items  ",
		Body:   "hello",
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if response.Status != 201 {
		t.Fatalf("expected status 201, got %d", response.Status)
	}

	if fake.gotRequest.Method != "POST" {
		t.Fatalf("expected method POST, got %q", fake.gotRequest.Method)
	}

	if fake.gotRequest.URL != "https://example.com/items" {
		t.Fatalf("expected trimmed URL, got %q", fake.gotRequest.URL)
	}

	if !reflect.DeepEqual(fake.gotRequest.Headers, map[string][]string{}) {
		t.Fatalf("expected headers to be initialized, got %#v", fake.gotRequest.Headers)
	}

}

func TestSendRequestUseCaseExecuteReturnsSenderError(t *testing.T) {
	wantErr := errors.New("send failed")
	useCase := SendRequestUseCase{
		Client: &fakeSender{err: wantErr},
	}

	_, err := useCase.Execute(domain.Request{URL: "https://example.com"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected sender error %v, got %v", wantErr, err)
	}
}