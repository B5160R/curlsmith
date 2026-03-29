package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInsertModePersistsAcrossKeys(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.focusIndex = buildFocusURL
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
	if updated.focusIndex != buildFocusURL {
		t.Fatalf("expected focus to stay on URL field in insert mode, got %d", updated.focusIndex)
	}
}

func TestNormalModeJMovesFocus(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.focusIndex = buildFocusURL
	m.insertMode = false

	updatedModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	updated := updatedModel.(model)

	if updated.focusIndex != buildFocusCollectionPath {
		t.Fatalf("expected j to move focus to collection path, got %d", updated.focusIndex)
	}
}

func TestEscapeLeavesInsertMode(t *testing.T) {
	m := NewModel(nil, nil, nil)
	m.focusIndex = buildFocusHeaders
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
