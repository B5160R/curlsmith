package tui

import (
	"errors"
	"testing"

	"github.com/b5160r/curlsmith/internal/domain"
	tea "github.com/charmbracelet/bubbletea"
)

// --- Fakes ---

type fakeExecutor struct {
	gotRequest domain.Request
	response   domain.Response
	err        error
}

func (f *fakeExecutor) Execute(req domain.Request) (domain.Response, error) {
	f.gotRequest = req
	return f.response, f.err
}

type fakeCollectionExecutor struct {
	gotPath    string
	collection domain.Collection
	err        error
}

func (f *fakeCollectionExecutor) Execute(path string) (domain.Collection, error) {
	f.gotPath = path
	return f.collection, f.err
}

type fakeCollectionSaver struct {
	gotPath    string
	gotRequest domain.Request
	collection domain.Collection
	err        error
}

func (f *fakeCollectionSaver) Execute(path string, req domain.Request) (domain.Collection, error) {
	f.gotPath = path
	f.gotRequest = req
	return f.collection, f.err
}

// --- Insert/Normal mode tests ---

func TestInsertModePersistsAcrossKeys(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.focusIndex = int(buildFocusURL)
	m.insertMode = false

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	updated := updatedModel.(model)
	if !updated.insertMode {
		t.Fatal("expected insert mode to turn on after pressing i")
	}

	updatedModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	updated = updatedModel.(model)
	if !updated.insertMode {
		t.Fatal("expected insert mode to stay on while typing")
	}
	if updated.focusIndex != int(buildFocusURL) {
		t.Fatalf("expected focus to stay on URL field in insert mode, got %d", updated.focusIndex)
	}
}

func TestNormalModeJMovesFocus(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.focusIndex = int(buildFocusURL)
	m.insertMode = false

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	updated := updatedModel.(model)

	if updated.focusIndex != int(buildFocusCollectionPath) {
		t.Fatalf("expected j to move focus to collection path, got %d", updated.focusIndex)
	}
}

func TestNormalModeKMovesFocusUp(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.focusIndex = int(buildFocusCollectionPath)
	m.insertMode = false

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	updated := updatedModel.(model)

	if updated.focusIndex != int(buildFocusURL) {
		t.Fatalf("expected k to move focus up to URL, got %d", updated.focusIndex)
	}
}

func TestEscapeLeavesInsertMode(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.focusIndex = int(buildFocusHeaders)
	m.insertMode = true

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	updated := updatedModel.(model)

	if updated.insertMode {
		t.Fatal("expected esc to leave insert mode")
	}
}

func TestModeLabelReflectsInsertState(t *testing.T) {
	m := NewModel(nil, nil, nil)

	if got := m.modeLabel(); got != "NORMAL" {
		t.Fatalf("expected NORMAL label, got %q", got)
	}

	m.insertMode = true
	if got := m.modeLabel(); got != "INSERT" {
		t.Fatalf("expected INSERT label, got %q", got)
	}
}

func TestQKeyQuitsInNormalMode(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.insertMode = false

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit command from q key in normal mode")
	}
}

func TestQKeyDoesNotQuitInInsertMode(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.focusIndex = int(buildFocusURL)
	m.insertMode = true

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	// In insert mode, q should be passed to the input widget, not quit.
	// The command should not be a quit command.
	if cmd != nil {
		// Execute the command to check it's not a quit msg
		msg := cmd()
		if _, ok := msg.(tea.QuitMsg); ok {
			t.Fatal("q in insert mode should not quit")
		}
	}
}

// --- Mode switching ---

func TestRightKeySwitchesToLoadMode(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.focusIndex = m.modeFocusIndex()
	m.insertMode = false

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	updated := updatedModel.(model)

	if updated.mode != loadMode {
		t.Fatalf("expected load mode, got %d", updated.mode)
	}
}

func TestLeftKeySwitchesToBuildMode(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.mode = loadMode
	m.focusIndex = m.modeFocusIndex()
	m.insertMode = false

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	updated := updatedModel.(model)

	if updated.mode != buildMode {
		t.Fatal("expected build mode after pressing h on mode tab")
	}
}

func TestMethodCyclesWithLeftRight(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.focusIndex = int(buildFocusMethod)
	m.methodIndex = 0
	m.insertMode = false

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	updated := updatedModel.(model)
	if updated.methodIndex != 1 {
		t.Fatalf("expected method index 1, got %d", updated.methodIndex)
	}

	updatedModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	updated = updatedModel.(model)
	if updated.methodIndex != 0 {
		t.Fatalf("expected method index back to 0, got %d", updated.methodIndex)
	}
}

// --- Send request (s key) ---

func TestSKeySendsRequest(t *testing.T) {
	executor := &fakeExecutor{response: domain.Response{Status: 200}}
	m := NewModel(executor, nil, nil)
	m.focusIndex = int(buildFocusURL)
	m.insertMode = false
	m.urlInput.SetValue("https://example.com")

	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	updated := updatedModel.(model)

	if !updated.sending {
		t.Fatal("expected sending flag to be true")
	}
	if cmd == nil {
		t.Fatal("expected a command to be returned")
	}
}

func TestSKeyWithInvalidHeadersShowsError(t *testing.T) {
	m := NewModel(&fakeExecutor{}, nil, nil)
	m.focusIndex = int(buildFocusURL)
	m.insertMode = false
	m.urlInput.SetValue("https://example.com")
	m.headersInput.SetValue("bad header no colon")

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	updated := updatedModel.(model)

	if updated.lastError == "" {
		t.Fatal("expected error when headers are invalid")
	}
	if updated.sending {
		t.Fatal("should not be in sending state on error")
	}
}

func TestSKeyIgnoredInInsertMode(t *testing.T) {
	m := NewModel(&fakeExecutor{}, nil, nil)
	m.focusIndex = int(buildFocusURL)
	m.insertMode = true
	m.urlInput.SetValue("https://example.com")

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	updated := updatedModel.(model)

	if updated.sending {
		t.Fatal("s in insert mode should not send")
	}
}

// --- Save request (w key) ---

func TestWKeySavesRequest(t *testing.T) {
	saver := &fakeCollectionSaver{collection: domain.Collection{Requests: []domain.Request{{URL: "https://example.com"}}}}
	m := NewModel(&fakeExecutor{}, nil, saver)
	m.focusIndex = int(buildFocusURL)
	m.insertMode = false
	m.urlInput.SetValue("https://example.com")

	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	updated := updatedModel.(model)

	if !updated.savingCollection {
		t.Fatal("expected savingCollection to be true")
	}
	if cmd == nil {
		t.Fatal("expected command from w key")
	}
}

// --- Load collection (r key) ---

func TestRKeyLoadsCollectionInLoadMode(t *testing.T) {
	loader := &fakeCollectionExecutor{collection: domain.Collection{
		Requests: []domain.Request{{Name: "test", URL: "http://localhost"}},
	}}
	m := NewModel(nil, loader, nil)
	m.mode = loadMode
	m.focusIndex = int(loadFocusPath)
	m.insertMode = false

	updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	updated := updatedModel.(model)

	if !updated.loadingCollection {
		t.Fatal("expected loadingCollection to be true")
	}
	if cmd == nil {
		t.Fatal("expected command from r key")
	}
}

func TestRKeyIgnoredInBuildMode(t *testing.T) {
	m := NewModel(nil, &fakeCollectionExecutor{}, nil)
	m.mode = buildMode
	m.insertMode = false

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	updated := updatedModel.(model)

	if updated.loadingCollection {
		t.Fatal("r should be ignored in build mode")
	}
}

// --- Message handling ---

func TestResponseMsgUpdatesState(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.sending = true
	m.width = 100
	m.height = 40

	resp := domain.Response{Status: 201, Body: []byte("created")}
	updatedModel, _ := m.Update(responseMsg{response: resp})
	updated := updatedModel.(model)

	if updated.sending {
		t.Fatal("sending should be false after response")
	}
	if updated.response.Status != 201 {
		t.Fatalf("expected status 201, got %d", updated.response.Status)
	}
	if updated.lastError != "" {
		t.Fatalf("expected no error, got %q", updated.lastError)
	}
}

func TestResponseMsgWithErrorSetsLastError(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.sending = true
	m.width = 100
	m.height = 40

	updatedModel, _ := m.Update(responseMsg{err: errors.New("timeout")})
	updated := updatedModel.(model)

	if updated.sending {
		t.Fatal("sending should be false after error response")
	}
	if updated.lastError != "timeout" {
		t.Fatalf("expected error 'timeout', got %q", updated.lastError)
	}
}

func TestCollectionMsgUpdatesState(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.loadingCollection = true
	m.width = 100
	m.height = 40

	col := domain.Collection{
		Requests: []domain.Request{{Name: "req1"}, {Name: "req2"}},
	}
	updatedModel, _ := m.Update(collectionMsg{path: "./test.json", collection: col})
	updated := updatedModel.(model)

	if updated.loadingCollection {
		t.Fatal("loadingCollection should be false")
	}
	if len(updated.collection.Requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(updated.collection.Requests))
	}
	if updated.selectedRequestIndex != 0 {
		t.Fatalf("expected selectedRequestIndex 0, got %d", updated.selectedRequestIndex)
	}
	if updated.loadedCollectionPath != "./test.json" {
		t.Fatalf("expected path ./test.json, got %q", updated.loadedCollectionPath)
	}
}

func TestCollectionMsgWithErrorClearsState(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.loadingCollection = true
	m.width = 100
	m.height = 40

	updatedModel, _ := m.Update(collectionMsg{err: errors.New("not found")})
	updated := updatedModel.(model)

	if updated.lastError != "not found" {
		t.Fatalf("expected error, got %q", updated.lastError)
	}
	if updated.selectedRequestIndex != -1 {
		t.Fatal("expected selectedRequestIndex to be -1")
	}
}

func TestCollectionSaveMsgUpdatesState(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.savingCollection = true
	m.width = 100
	m.height = 40
	m.nameInput.SetValue("my req")
	m.urlInput.SetValue("https://example.com")

	col := domain.Collection{
		Requests: []domain.Request{{Name: "my req", URL: "https://example.com"}},
	}
	updatedModel, _ := m.Update(collectionSaveMsg{path: "./col.json", collection: col})
	updated := updatedModel.(model)

	if updated.savingCollection {
		t.Fatal("savingCollection should be false after save")
	}
	if updated.lastError != "" {
		t.Fatalf("expected no error, got %q", updated.lastError)
	}
	if updated.selectedRequestIndex != 0 {
		t.Fatalf("expected selectedRequestIndex 0, got %d", updated.selectedRequestIndex)
	}
}

// --- Window resize ---

func TestWindowSizeMsgResizes(t *testing.T) {
	m := NewModel(nil, nil, nil)

	updatedModel, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	updated := updatedModel.(model)

	if updated.width != 120 || updated.height != 40 {
		t.Fatalf("expected 120x40, got %dx%d", updated.width, updated.height)
	}
}

// --- Load request navigation ---

func TestUpDownNavigatesRequests(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.mode = loadMode
	m.focusIndex = int(loadFocusRequests)
	m.insertMode = false
	m.collection = domain.Collection{
		Requests: []domain.Request{{Name: "a"}, {Name: "b"}, {Name: "c"}},
	}
	m.selectedRequestIndex = 0

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	updated := updatedModel.(model)
	if updated.selectedRequestIndex != 1 {
		t.Fatalf("expected index 1, got %d", updated.selectedRequestIndex)
	}

	updatedModel, _ = updated.Update(tea.KeyMsg{Type: tea.KeyUp})
	updated = updatedModel.(model)
	if updated.selectedRequestIndex != 0 {
		t.Fatalf("expected index 0, got %d", updated.selectedRequestIndex)
	}
}

func TestEnterLoadsRequestIntoForge(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.mode = loadMode
	m.focusIndex = int(loadFocusRequests)
	m.insertMode = false
	m.collection = domain.Collection{
		Requests: []domain.Request{{
			Name:    "test request",
			Method:  "POST",
			URL:     "https://example.com/test",
			Headers: map[string][]string{"X-Key": {"val"}},
			Body:    "body content",
		}},
	}
	m.selectedRequestIndex = 0

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	updated := updatedModel.(model)

	if updated.mode != buildMode {
		t.Fatal("expected mode to switch to build")
	}
	if updated.urlInput.Value() != "https://example.com/test" {
		t.Fatalf("expected URL loaded, got %q", updated.urlInput.Value())
	}
	if updated.nameInput.Value() != "test request" {
		t.Fatalf("expected name loaded, got %q", updated.nameInput.Value())
	}
	if updated.methodIndex != 1 { // POST
		t.Fatalf("expected method index 1 (POST), got %d", updated.methodIndex)
	}
}

// --- buildRequest ---

func TestBuildRequestReturnsValidRequest(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.nameInput.SetValue("test")
	m.urlInput.SetValue("https://example.com")
	m.methodIndex = 1 // POST
	m.headersInput.SetValue("Content-Type: application/json")
	m.bodyInput.SetValue(`{"key":"value"}`)

	req, err := m.buildRequest()
	if err != nil {
		t.Fatalf("buildRequest returned error: %v", err)
	}
	if req.Name != "test" {
		t.Fatalf("expected name 'test', got %q", req.Name)
	}
	if req.Method != "POST" {
		t.Fatalf("expected POST, got %q", req.Method)
	}
	if req.URL != "https://example.com" {
		t.Fatalf("expected URL, got %q", req.URL)
	}
	if req.Headers["Content-Type"][0] != "application/json" {
		t.Fatalf("expected Content-Type header, got %v", req.Headers)
	}
}

func TestBuildRequestWithInvalidHeadersReturnsError(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.urlInput.SetValue("https://example.com")
	m.headersInput.SetValue("no-colon-here")

	_, err := m.buildRequest()
	if err == nil {
		t.Fatal("expected error for invalid header format")
	}
}

// --- loadSelectedRequestIntoForge ---

func TestLoadSelectedRequestIntoForgeWithNoSelection(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.selectedRequestIndex = -1

	err := m.loadSelectedRequestIntoForge()
	if err == nil {
		t.Fatal("expected error when no request selected")
	}
}

func TestLoadSelectedRequestIntoForge(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.collection = domain.Collection{
		Requests: []domain.Request{{
			Name:   "loaded",
			Method: "DELETE",
			URL:    "https://api.test/item/1",
			Body:   "confirm",
		}},
	}
	m.selectedRequestIndex = 0

	err := m.loadSelectedRequestIntoForge()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.nameInput.Value() != "loaded" {
		t.Fatalf("expected name 'loaded', got %q", m.nameInput.Value())
	}
	if m.methodIndex != 4 { // DELETE
		t.Fatalf("expected DELETE index 4, got %d", m.methodIndex)
	}
}

// --- findSavedRequestIndex ---

func TestFindSavedRequestIndexByName(t *testing.T) {
	requests := []domain.Request{
		{Name: "first", URL: "http://a.com"},
		{Name: "second", URL: "http://b.com"},
	}
	if got := findSavedRequestIndex(requests, "second", ""); got != 1 {
		t.Fatalf("expected 1, got %d", got)
	}
}

func TestFindSavedRequestIndexByURL(t *testing.T) {
	requests := []domain.Request{
		{Name: "first", URL: "http://a.com"},
		{Name: "second", URL: "http://b.com"},
	}
	if got := findSavedRequestIndex(requests, "", "http://b.com"); got != 1 {
		t.Fatalf("expected 1, got %d", got)
	}
}

func TestFindSavedRequestIndexNotFound(t *testing.T) {
	requests := []domain.Request{{Name: "only", URL: "http://a.com"}}
	if got := findSavedRequestIndex(requests, "other", "http://other.com"); got != -1 {
		t.Fatalf("expected -1, got %d", got)
	}
}

// --- View ---

func TestViewBeforeResize(t *testing.T) {
	m := NewModel(nil, nil, nil)
	got := m.View()
	if got != "Heating the forge..." {
		t.Fatalf("expected loading message, got %q", got)
	}
}

func TestViewAfterResize(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.width = 120
	m.height = 40
	m.resize()

	got := m.View()
	if got == "Heating the forge..." {
		t.Fatal("expected full view after resize")
	}
}
