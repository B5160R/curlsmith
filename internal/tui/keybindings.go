package tui

import (
	"fmt"
	"strings"

	"github.com/b5160r/curlsmith/internal/domain"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) handleKeyMsg(msg tea.KeyMsg) (model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "i", "a", "o":
		if !m.insertMode && m.focusSupportsInsertMode() {
			m.insertMode = true
			m.syncFocus()
			return m, nil
		}
	case "esc":
		if m.insertMode {
			m.insertMode = false
			m.syncFocus()
			return m, nil
		}
	case "q":
		if !m.insertMode {
			return m, tea.Quit
		}
	case "j":
		if !m.insertMode {
			if m.mode == loadMode && m.focusIndex == int(loadFocusRequests) && len(m.collection.Requests) > 0 {
				m.selectedRequestIndex = (m.selectedRequestIndex + 1) % len(m.collection.Requests)
				return m, nil
			}
			m.focusIndex = (m.focusIndex + 1) % m.focusCount()
			m.syncFocus()
			return m, nil
		}
	case "k":
		if !m.insertMode {
			if m.mode == loadMode && m.focusIndex == int(loadFocusRequests) && len(m.collection.Requests) > 0 {
				m.selectedRequestIndex = (m.selectedRequestIndex + len(m.collection.Requests) - 1) % len(m.collection.Requests)
				return m, nil
			}
			m.focusIndex = (m.focusIndex + m.focusCount() - 1) % m.focusCount()
			m.syncFocus()
			return m, nil
		}
	case "left", "h":
		if m.insertMode {
			break
		}
		if m.focusIndex == m.modeFocusIndex() {
			m.setMode(buildMode)
			return m, nil
		}
		if m.mode == buildMode && m.focusIndex == int(buildFocusMethod) {
			m.methodIndex = (m.methodIndex + len(methods) - 1) % len(methods)
			return m, nil
		}
	case "right", "l":
		if m.insertMode {
			break
		}
		if m.focusIndex == m.modeFocusIndex() {
			m.setMode(loadMode)
			return m, nil
		}
		if m.mode == buildMode && m.focusIndex == int(buildFocusMethod) {
			m.methodIndex = (m.methodIndex + 1) % len(methods)
			return m, nil
		}
	case "s":
		if m.insertMode {
			break
		}
		if m.mode != buildMode || m.sending || m.savingCollection {
			return m, nil
		}

		req, err := m.buildRequest()
		if err != nil {
			m.lastError = err.Error()
			m.lastInfo = ""
			m.refreshResponseViewport()
			return m, nil
		}

		m.sending = true
		m.lastError = ""
		m.lastInfo = ""
		m.refreshResponseViewport()
		return m, sendRequestCmd(m.executor, req)
	case "w":
		if m.insertMode {
			break
		}
		if m.mode != buildMode || m.savingCollection || m.sending {
			return m, nil
		}

		req, err := m.buildRequest()
		if err != nil {
			m.lastError = err.Error()
			m.lastInfo = ""
			m.refreshResponseViewport()
			return m, nil
		}

		m.savingCollection = true
		m.lastError = ""
		m.lastInfo = ""
		m.refreshResponseViewport()
		return m, saveCollectionCmd(m.collectionSaver, m.collectionPathInput.Value(), req)
	case "r":
		if m.insertMode {
			break
		}
		if m.mode != loadMode || m.loadingCollection {
			return m, nil
		}
		m.loadingCollection = true
		m.lastError = ""
		m.lastInfo = ""
		m.refreshResponseViewport()
		return m, loadCollectionCmd(m.collectionExecutor, m.collectionPathInput.Value())
	case "up":
		if !m.insertMode && m.mode == loadMode && m.focusIndex == int(loadFocusRequests) && len(m.collection.Requests) > 0 {
			m.selectedRequestIndex = (m.selectedRequestIndex + len(m.collection.Requests) - 1) % len(m.collection.Requests)
			return m, nil
		}
	case "down":
		if !m.insertMode && m.mode == loadMode && m.focusIndex == int(loadFocusRequests) && len(m.collection.Requests) > 0 {
			m.selectedRequestIndex = (m.selectedRequestIndex + 1) % len(m.collection.Requests)
			return m, nil
		}
	case "enter":
		if m.mode == loadMode && m.focusIndex == int(loadFocusRequests) {
			if err := m.loadSelectedRequestIntoForge(); err != nil {
				m.lastError = err.Error()
				m.lastInfo = ""
				m.refreshResponseViewport()
				return m, nil
			}
			m.lastError = ""
			m.lastInfo = "Request pulled from collection into the forge bench"
			m.setMode(buildMode)
			m.focusIndex = int(buildFocusURL)
			m.syncFocus()
			return m, nil
		}
	}

	// Delegate to focused widget.
	if m.mode == buildMode {
		switch m.focusIndex {
		case int(buildFocusName):
			m.nameInput, cmd = m.nameInput.Update(msg)
			return m, cmd
		case int(buildFocusURL):
			m.urlInput, cmd = m.urlInput.Update(msg)
			return m, cmd
		case int(buildFocusCollectionPath):
			m.collectionPathInput, cmd = m.collectionPathInput.Update(msg)
			return m, cmd
		case int(buildFocusHeaders):
			m.headersInput, cmd = m.headersInput.Update(msg)
			return m, cmd
		case int(buildFocusBody):
			m.bodyInput, cmd = m.bodyInput.Update(msg)
			return m, cmd
		case int(buildFocusResponse):
			m.responseViewport, cmd = m.responseViewport.Update(msg)
			return m, cmd
		}
	}

	if m.mode == loadMode {
		switch m.focusIndex {
		case int(loadFocusPath):
			m.collectionPathInput, cmd = m.collectionPathInput.Update(msg)
			return m, cmd
		case int(loadFocusResponse):
			m.responseViewport, cmd = m.responseViewport.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func findSavedRequestIndex(requests []domain.Request, name, url string) int {
	name = strings.TrimSpace(name)
	url = strings.TrimSpace(url)

	if name != "" {
		for index, request := range requests {
			if strings.EqualFold(strings.TrimSpace(request.Name), name) {
				return index
			}
		}
	}

	for index, request := range requests {
		if strings.TrimSpace(request.URL) == url {
			return index
		}
	}

	return -1
}

func (m *model) loadSelectedRequestIntoForge() error {
	if len(m.collection.Requests) == 0 || m.selectedRequestIndex < 0 || m.selectedRequestIndex >= len(m.collection.Requests) {
		return fmt.Errorf("no stored request is selected")
	}

	request := m.collection.Requests[m.selectedRequestIndex]
	m.nameInput.SetValue(request.Name)
	m.urlInput.SetValue(request.URL)
	m.headersInput.SetValue(renderHeadersInput(request.Headers))
	m.bodyInput.SetValue(request.Body)
	m.methodIndex = findMethodIndex(request.Method)
	return nil
}
