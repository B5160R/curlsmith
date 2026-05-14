package tui

// viewMode represents the top-level TUI mode.
type viewMode int

const (
	buildMode viewMode = iota
	loadMode
)

// buildFocus represents a focusable field in build mode.
type buildFocus int

const (
	buildFocusMode buildFocus = iota
	buildFocusMethod
	buildFocusName
	buildFocusURL
	buildFocusCollectionPath
	buildFocusHeaders
	buildFocusBody
	buildFocusResponse
)

// loadFocus represents a focusable field in load mode.
type loadFocus int

const (
	loadFocusMode loadFocus = iota
	loadFocusPath
	loadFocusRequests
	loadFocusResponse
)

func (m *model) syncFocus() {
	m.nameInput.Blur()
	m.urlInput.Blur()
	m.collectionPathInput.Blur()
	m.headersInput.Blur()
	m.bodyInput.Blur()

	if !m.focusSupportsInsertMode() {
		m.insertMode = false
	}

	if !m.insertMode {
		return
	}

	if m.mode == buildMode {
		switch m.focusIndex {
		case int(buildFocusName):
			m.nameInput.Focus()
		case int(buildFocusURL):
			m.urlInput.Focus()
		case int(buildFocusCollectionPath):
			m.collectionPathInput.Focus()
		case int(buildFocusHeaders):
			m.headersInput.Focus()
		case int(buildFocusBody):
			m.bodyInput.Focus()
		}
		return
	}

	if m.focusIndex == int(loadFocusPath) {
		m.collectionPathInput.Focus()
	}
}

func (m model) focusSupportsInsertMode() bool {
	if m.mode == buildMode {
		switch m.focusIndex {
		case int(buildFocusName), int(buildFocusURL), int(buildFocusCollectionPath), int(buildFocusHeaders), int(buildFocusBody):
			return true
		}
		return false
	}

	return m.focusIndex == int(loadFocusPath)
}

func (m model) focusCount() int {
	if m.mode == buildMode {
		return int(buildFocusResponse) + 1
	}
	return int(loadFocusResponse) + 1
}

func (m model) modeFocusIndex() int {
	return 0
}

func (m *model) setMode(mode viewMode) {
	if m.mode == mode {
		return
	}

	m.mode = mode
	if mode == buildMode {
		if m.focusIndex == int(loadFocusPath) {
			m.focusIndex = int(buildFocusCollectionPath)
		} else if m.focusIndex == int(loadFocusRequests) {
			m.focusIndex = int(buildFocusName)
		} else {
			m.focusIndex = min(m.focusIndex, m.focusCount()-1)
		}
	} else if m.focusIndex > int(loadFocusResponse) {
		m.focusIndex = int(loadFocusPath)
	}
	if mode == loadMode && m.focusIndex == int(buildFocusMethod) {
		m.focusIndex = int(loadFocusPath)
	}
	if mode == loadMode && m.focusIndex == int(buildFocusCollectionPath) {
		m.focusIndex = int(loadFocusPath)
	}
	m.syncFocus()
}
