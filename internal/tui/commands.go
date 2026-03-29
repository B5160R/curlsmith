package tui

import (
	"fmt"
	"strings"

	"github.com/b5160r/curlsmith/internal/domain"
	tea "github.com/charmbracelet/bubbletea"
)

type responseMsg struct {
	response domain.Response
	err      error
}

type collectionMsg struct {
	path       string
	collection domain.Collection
	err        error
}

type collectionSaveMsg struct {
	path       string
	collection domain.Collection
	err        error
}

func sendRequestCmd(executor RequestExecutor, req domain.Request) tea.Cmd {
	return func() tea.Msg {
		if executor == nil {
			return responseMsg{err: fmt.Errorf("request executor is not configured")}
		}

		response, err := executor.Execute(req)
		return responseMsg{response: response, err: err}
	}
}

func loadCollectionCmd(executor CollectionExecutor, path string) tea.Cmd {
	return func() tea.Msg {
		if executor == nil {
			return collectionMsg{err: fmt.Errorf("collection executor is not configured")}
		}

		collection, err := executor.Execute(path)
		return collectionMsg{path: strings.TrimSpace(path), collection: collection, err: err}
	}
}

func saveCollectionCmd(executor CollectionSaveExecutor, path string, req domain.Request) tea.Cmd {
	return func() tea.Msg {
		if executor == nil {
			return collectionSaveMsg{err: fmt.Errorf("collection saver is not configured")}
		}

		collection, err := executor.Execute(path, req)
		return collectionSaveMsg{path: strings.TrimSpace(path), collection: collection, err: err}
	}
}
