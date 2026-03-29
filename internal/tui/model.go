package tui

import (
	"fmt"
	"strings"

	"github.com/b5160r/curlsmith/internal/domain"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type RequestExecutor interface {
	Execute(req domain.Request) (domain.Response, error)
}

type CollectionExecutor interface {
	Execute(path string) (domain.Collection, error)
}

type CollectionSaveExecutor interface {
	Execute(path string, req domain.Request) (domain.Collection, error)
}

type viewMode int

const (
	buildMode viewMode = iota
	loadMode
)

const (
	buildFocusMode = iota
	buildFocusMethod
	buildFocusName
	buildFocusURL
	buildFocusCollectionPath
	buildFocusHeaders
	buildFocusBody
	buildFocusResponse
)

const (
	loadFocusMode = iota
	loadFocusPath
	loadFocusRequests
	loadFocusResponse
)

type model struct {
	executor             RequestExecutor
	collectionExecutor   CollectionExecutor
	collectionSaver      CollectionSaveExecutor
	width                int
	height               int
	mode                 viewMode
	insertMode           bool
	focusIndex           int
	methodIndex          int
	selectedRequestIndex int
	sending              bool
	loadingCollection    bool
	savingCollection     bool
	lastError            string
	lastInfo             string
	response             domain.Response
	collection           domain.Collection
	loadedCollectionPath string
	responseViewport     viewport.Model
	nameInput            textinput.Model
	urlInput             textinput.Model
	collectionPathInput  textinput.Model
	headersInput         textarea.Model
	bodyInput            textarea.Model
	styles               styles
}

var methods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

func NewModel(executor RequestExecutor, collectionExecutor CollectionExecutor, collectionSaver CollectionSaveExecutor) model {
	nameInput := textinput.New()
	nameInput.Placeholder = "request name"
	nameInput.Prompt = ""

	urlInput := textinput.New()
	urlInput.Placeholder = "https://api.blacksmith.test/orders"
	urlInput.Prompt = ""
	urlInput.Focus()

	collectionPathInput := textinput.New()
	collectionPathInput.Prompt = ""
	collectionPathInput.Placeholder = "./collections/armory.json"
	collectionPathInput.SetValue("./curlsmith.collection.json")

	headersInput := textarea.New()
	headersInput.Placeholder = "Content-Type: application/json\nX-Trace: forge-01"
	headersInput.ShowLineNumbers = false
	headersInput.SetHeight(5)
	headersInput.Blur()

	bodyInput := textarea.New()
	bodyInput.Placeholder = "{\n  \"steel\": \"damascus\"\n}"
	bodyInput.ShowLineNumbers = false
	bodyInput.SetHeight(10)
	bodyInput.Blur()

	responseViewport := viewport.New(40, 18)
	responseViewport.SetContent("No request forged yet. Fill the forge and press s to send it downrange.")

	model := model{
		executor:             executor,
		collectionExecutor:   collectionExecutor,
		collectionSaver:      collectionSaver,
		mode:                 buildMode,
		insertMode:           false,
		focusIndex:           buildFocusURL,
		methodIndex:          0,
		selectedRequestIndex: -1,
		responseViewport:     responseViewport,
		nameInput:            nameInput,
		urlInput:             urlInput,
		collectionPathInput:  collectionPathInput,
		headersInput:         headersInput,
		bodyInput:            bodyInput,
		styles:               newStyles(),
	}

	model.syncFocus()
	return model
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resize()
		m.refreshResponseViewport()
		return m, nil

	case responseMsg:
		m.sending = false
		m.lastError = ""
		m.lastInfo = ""
		if msg.err != nil {
			m.lastError = msg.err.Error()
			m.response = domain.Response{}
		} else {
			m.response = msg.response
		}
		m.refreshResponseViewport()
		return m, nil

	case collectionMsg:
		m.loadingCollection = false
		if msg.err != nil {
			m.lastError = msg.err.Error()
			m.lastInfo = ""
			m.collection = domain.Collection{}
			m.loadedCollectionPath = ""
			m.selectedRequestIndex = -1
		} else {
			m.lastError = ""
			m.lastInfo = fmt.Sprintf("Loaded %d request(s) from %s", len(msg.collection.Requests), msg.path)
			m.collection = msg.collection
			m.loadedCollectionPath = msg.path
			if len(msg.collection.Requests) > 0 {
				m.selectedRequestIndex = 0
			} else {
				m.selectedRequestIndex = -1
			}
		}
		m.refreshResponseViewport()
		return m, nil

	case collectionSaveMsg:
		m.savingCollection = false
		if msg.err != nil {
			m.lastError = msg.err.Error()
			m.lastInfo = ""
		} else {
			m.lastError = ""
			m.lastInfo = fmt.Sprintf("Saved request to %s", msg.path)
			m.collection = msg.collection
			m.loadedCollectionPath = msg.path
			m.selectedRequestIndex = findSavedRequestIndex(msg.collection.Requests, strings.TrimSpace(m.nameInput.Value()), strings.TrimSpace(m.urlInput.Value()))
		}
		m.refreshResponseViewport()
		return m, nil

	case tea.KeyMsg:
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
				if m.mode == loadMode && m.focusIndex == loadFocusRequests && len(m.collection.Requests) > 0 {
					m.selectedRequestIndex = (m.selectedRequestIndex + 1) % len(m.collection.Requests)
					return m, nil
				}
				m.focusIndex = (m.focusIndex + 1) % m.focusCount()
				m.syncFocus()
				return m, nil
			}
		case "k":
			if !m.insertMode {
				if m.mode == loadMode && m.focusIndex == loadFocusRequests && len(m.collection.Requests) > 0 {
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
			if m.mode == buildMode && m.focusIndex == buildFocusMethod {
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
			if m.mode == buildMode && m.focusIndex == buildFocusMethod {
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
			if !m.insertMode && m.mode == loadMode && m.focusIndex == loadFocusRequests && len(m.collection.Requests) > 0 {
				m.selectedRequestIndex = (m.selectedRequestIndex + len(m.collection.Requests) - 1) % len(m.collection.Requests)
				return m, nil
			}
		case "down":
			if !m.insertMode && m.mode == loadMode && m.focusIndex == loadFocusRequests && len(m.collection.Requests) > 0 {
				m.selectedRequestIndex = (m.selectedRequestIndex + 1) % len(m.collection.Requests)
				return m, nil
			}
		case "enter":
			if m.mode == loadMode && m.focusIndex == loadFocusRequests {
				if err := m.loadSelectedRequestIntoForge(); err != nil {
					m.lastError = err.Error()
					m.lastInfo = ""
					m.refreshResponseViewport()
					return m, nil
				}
				m.lastError = ""
				m.lastInfo = "Request pulled from collection into the forge bench"
				m.setMode(buildMode)
				m.focusIndex = buildFocusURL
				m.syncFocus()
				return m, nil
			}
		}
	}

	if m.mode == buildMode {
		switch m.focusIndex {
		case buildFocusName:
			m.nameInput, cmd = m.nameInput.Update(msg)
			return m, cmd
		case buildFocusURL:
			m.urlInput, cmd = m.urlInput.Update(msg)
			return m, cmd
		case buildFocusCollectionPath:
			m.collectionPathInput, cmd = m.collectionPathInput.Update(msg)
			return m, cmd
		case buildFocusHeaders:
			m.headersInput, cmd = m.headersInput.Update(msg)
			return m, cmd
		case buildFocusBody:
			m.bodyInput, cmd = m.bodyInput.Update(msg)
			return m, cmd
		case buildFocusResponse:
			m.responseViewport, cmd = m.responseViewport.Update(msg)
			return m, cmd
		}
	}

	if m.mode == loadMode {
		switch m.focusIndex {
		case loadFocusPath:
			m.collectionPathInput, cmd = m.collectionPathInput.Update(msg)
			return m, cmd
		case loadFocusResponse:
			m.responseViewport, cmd = m.responseViewport.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Heating the forge..."
	}

	titleRow := lipgloss.JoinHorizontal(lipgloss.Center, m.styles.header.Render("Curlsmith"), " ", m.styles.modeBadge(m.insertMode).Render(m.modeLabel()))
	header := titleRow + "\n" + m.styles.subheader.Render("Forge a request like a smith at the anvil: shape it, temper it, load it, send it.")
	left := m.renderForgePanel()
	right := m.renderResponsePanel()
	footer := m.styles.footer.Render(m.footerText())

	body := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

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
		case buildFocusName:
			m.nameInput.Focus()
		case buildFocusURL:
			m.urlInput.Focus()
		case buildFocusCollectionPath:
			m.collectionPathInput.Focus()
		case buildFocusHeaders:
			m.headersInput.Focus()
		case buildFocusBody:
			m.bodyInput.Focus()
		}
		return
	}

	if m.focusIndex == loadFocusPath {
		m.collectionPathInput.Focus()
	}
}

func (m *model) resize() {
	leftWidth := max(44, m.width*48/100)
	rightWidth := max(38, m.width-leftWidth-2)
	leftInnerWidth := max(24, leftWidth-8)
	rightInnerWidth := max(24, rightWidth-6)

	m.nameInput.Width = leftInnerWidth
	m.urlInput.Width = leftInnerWidth
	m.collectionPathInput.Width = leftInnerWidth
	m.headersInput.SetWidth(leftInnerWidth)
	m.bodyInput.SetWidth(leftInnerWidth)

	headersHeight := max(4, min(7, m.height/5))
	bodyHeight := max(7, m.height/4)
	m.headersInput.SetHeight(headersHeight)
	m.bodyInput.SetHeight(bodyHeight)

	responseHeight := max(10, m.height-11)
	m.responseViewport.Width = rightInnerWidth
	m.responseViewport.Height = responseHeight
}

func (m *model) buildRequest() (domain.Request, error) {
	headers, err := parseHeaders(m.headersInput.Value())
	if err != nil {
		return domain.Request{}, err
	}

	return domain.Request{
		Name:    strings.TrimSpace(m.nameInput.Value()),
		Method:  methods[m.methodIndex],
		URL:     strings.TrimSpace(m.urlInput.Value()),
		Headers: headers,
		Body:    m.bodyInput.Value(),
	}, nil
}

func (m *model) setMode(mode viewMode) {
	if m.mode == mode {
		return
	}

	m.mode = mode
	if mode == buildMode {
		if m.focusIndex == loadFocusPath {
			m.focusIndex = buildFocusCollectionPath
		} else if m.focusIndex == loadFocusRequests {
			m.focusIndex = buildFocusName
		} else {
			m.focusIndex = min(m.focusIndex, m.focusCount()-1)
		}
	} else if m.focusIndex > loadFocusResponse {
		m.focusIndex = loadFocusPath
	}
	if mode == loadMode && m.focusIndex == buildFocusMethod {
		m.focusIndex = loadFocusPath
	}
	if mode == loadMode && m.focusIndex == buildFocusCollectionPath {
		m.focusIndex = loadFocusPath
	}
	m.syncFocus()
}

func (m model) focusSupportsInsertMode() bool {
	if m.mode == buildMode {
		switch m.focusIndex {
		case buildFocusName, buildFocusURL, buildFocusCollectionPath, buildFocusHeaders, buildFocusBody:
			return true
		}
		return false
	}

	return m.focusIndex == loadFocusPath
}

func (m model) focusCount() int {
	if m.mode == buildMode {
		return 8
	}
	return 4
}

func (m model) modeFocusIndex() int {
	return 0
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

func (m *model) refreshResponseViewport() {
	m.responseViewport.SetContent(m.responseContent())
	m.responseViewport.GotoTop()
}
