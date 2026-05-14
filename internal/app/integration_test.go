package app_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/b5160r/curlsmith/internal/app"
	"github.com/b5160r/curlsmith/internal/domain"
	httpinfra "github.com/b5160r/curlsmith/internal/infra/http"
	"github.com/b5160r/curlsmith/internal/infra/storage"
)

func TestIntegrationSaveLoadAndSend(t *testing.T) {
	// Set up a test HTTP server.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/items" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer server.Close()

	collectionPath := filepath.Join(t.TempDir(), "integration.json")
	store := storage.FileStore{}

	// 1. Save a request to the collection.
	saveUC := &app.SaveRequestToCollectionUseCase{Store: store}
	col, err := saveUC.Execute(collectionPath, domain.Request{
		Name:    "list items",
		Method:  "GET",
		URL:     "{{base_url}}/api/items",
		Headers: map[string][]string{"Accept": {"application/json"}},
	})
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}
	if len(col.Requests) != 1 {
		t.Fatalf("expected 1 request in collection, got %d", len(col.Requests))
	}

	// 2. Save collection variables manually.
	col.Variables = map[string]string{"base_url": server.URL}
	if err := store.Save(collectionPath, col); err != nil {
		t.Fatalf("failed to save variables: %v", err)
	}

	// 3. Load the collection back.
	loadUC := &app.LoadCollectionUseCase{Loader: store}
	loaded, err := loadUC.Execute(collectionPath)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(loaded.Requests) != 1 {
		t.Fatalf("expected 1 request after load, got %d", len(loaded.Requests))
	}
	if loaded.Variables["base_url"] != server.URL {
		t.Fatalf("expected base_url variable, got %v", loaded.Variables)
	}

	// 4. Send the request with variable resolution.
	sendUC := &app.SendRequestUseCase{
		Client: &httpinfra.Client{HTTPClient: server.Client()},
		Vars:   loaded.Variables,
	}
	resp, err := sendUC.Execute(loaded.Requests[0])
	if err != nil {
		t.Fatalf("send failed: %v", err)
	}
	if resp.Status != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Status)
	}
	if string(resp.Body) != `{"items":[]}` {
		t.Fatalf("unexpected body: %q", string(resp.Body))
	}
}
