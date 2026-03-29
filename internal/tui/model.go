package tui

import (
	"fmt"
	"sort"
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
	responseViewport.SetContent("No request forged yet. Fill the forge and press ctrl+s to send it downrange.")

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
		case "ctrl+s":
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
		case "ctrl+w":
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
		case "ctrl+r":
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

func (m *model) renderForgePanel() string {
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		m.styles.sectionTitle.Render("Forge Bench"),
		m.renderModeTabs(),
	)

	if m.mode == buildMode {
		methodField := m.styles.methodField(m.focusIndex == buildFocusMethod).Render(methods[m.methodIndex])
		nameField := m.styles.inputWrap(m.focusIndex == buildFocusName, m.insertMode).Render(m.nameInput.View())
		urlField := m.styles.inputWrap(m.focusIndex == buildFocusURL, m.insertMode).Render(m.urlInput.View())
		collectionPathField := m.styles.inputWrap(m.focusIndex == buildFocusCollectionPath, m.insertMode).Render(m.collectionPathInput.View())
		headersField := m.styles.inputWrap(m.focusIndex == buildFocusHeaders, m.insertMode).Render(m.headersInput.View())
		bodyField := m.styles.inputWrap(m.focusIndex == buildFocusBody, m.insertMode).Render(m.bodyInput.View())

		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			m.styles.modeHint.Render("Build mode shapes a request by hand. Press ctrl+s to send it or ctrl+w to write it into a collection file."),
			m.renderField("Method", methodField),
			m.renderField("Name", nameField),
			m.renderField("URL", urlField),
			m.renderField("Collection Path", collectionPathField),
			m.renderField("Headers", headersField),
			m.renderField("Body", bodyField),
		)
	} else {
		pathField := m.styles.inputWrap(m.focusIndex == loadFocusPath, m.insertMode).Render(m.collectionPathInput.View())
		requestField := m.styles.requestListWrap(m.focusIndex == loadFocusRequests).Render(m.renderStoredRequests())

		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			m.styles.modeHint.Render("Load mode pulls a collection off the rack. Press ctrl+r to read it, then enter on a request to temper it into the forge."),
			m.renderField("Collection Path", pathField),
			m.renderField("Collection Notes", m.styles.summaryBox.Render(m.renderCollectionSummary())),
			m.renderField("Stored Requests", requestField),
		)
	}

	return m.styles.panel.Render(content)
}

func (m model) renderModeTabs() string {
	buildFocused := m.focusIndex == m.modeFocusIndex() && m.mode == buildMode
	loadFocused := m.focusIndex == m.modeFocusIndex() && m.mode == loadMode

	buildTab := m.styles.tab(m.mode == buildMode, buildFocused).Render("Build Request")
	loadTab := m.styles.tab(m.mode == loadMode, loadFocused).Render("Load Collection")
	return lipgloss.JoinHorizontal(lipgloss.Left, buildTab, " ", loadTab)
}

func (m model) renderCollectionSummary() string {
	if m.loadingCollection {
		return "Heating tongs around collection file..."
	}

	if m.loadedCollectionPath == "" {
		return strings.Join([]string{
			"No collection loaded yet.",
			"Use the path field above and press ctrl+r.",
		}, "\n")
	}

	lines := []string{
		fmt.Sprintf("Path: %s", m.loadedCollectionPath),
		fmt.Sprintf("Requests: %d", len(m.collection.Requests)),
		fmt.Sprintf("Variables: %d", len(m.collection.Variables)),
	}

	if len(m.collection.Variables) > 0 {
		lines = append(lines, "", "Variables")
		lines = append(lines, formatVariables(m.collection.Variables)...)
	}

	return strings.Join(lines, "\n")
}

func (m model) renderStoredRequests() string {
	if len(m.collection.Requests) == 0 {
		return "(no requests in collection)"
	}

	lines := make([]string, 0, len(m.collection.Requests)*2)
	for index, request := range m.collection.Requests {
		marker := "  "
		if index == m.selectedRequestIndex {
			marker = "> "
		}

		name := request.Name
		if strings.TrimSpace(name) == "" {
			name = "unnamed request"
		}

		lines = append(lines, fmt.Sprintf("%s[%s] %s", marker, strings.ToUpper(request.Method), name))
		lines = append(lines, fmt.Sprintf("   %s", request.URL))
	}

	return strings.Join(lines, "\n")
}

func (m model) renderResponsePanel() string {
	status := "Awaiting first spark"
	if m.sending {
		status = "Hammering request through the fire..."
	} else if m.savingCollection {
		status = "Quenching request into collection steel..."
	} else if m.loadingCollection {
		status = "Pulling a collection from the rack..."
	} else if m.lastError != "" {
		status = "Forge cracked"
	} else if m.lastInfo != "" {
		status = m.lastInfo
	} else if m.response.Status != 0 {
		status = fmt.Sprintf("Status %d in %s", m.response.Status, m.response.Duration)
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		m.styles.sectionTitle.Render("Anvil Ledger"),
		m.styles.status.Render(status),
		m.styles.responseWrap(m.responseFocused()).Render(m.responseViewport.View()),
	)

	return m.styles.panel.Render(content)
}

func (m model) renderField(label, value string) string {
	return lipgloss.JoinVertical(lipgloss.Left, m.styles.label.Render(label), value)
}

func (m *model) refreshResponseViewport() {
	m.responseViewport.SetContent(m.responseContent())
	m.responseViewport.GotoTop()
}

func (m model) responseContent() string {
	if m.sending {
		return "The billet is in the fire. When the response returns, it will be etched here."
	}

	if m.savingCollection {
		return "The smith is riveting the request into a collection file."
	}

	if m.loadingCollection {
		return "The rackman is hauling a collection into the forge hall."
	}

	if m.lastError != "" {
		return "Operation failed\n\n" + m.lastError
	}

	if m.response.Status == 0 {
		if m.mode == loadMode {
			return strings.Join([]string{
				"Load notes",
				"",
				"- Move to Load Collection with the tabs at the top of the Forge Bench.",
				"- Enter a collection file path and press ctrl+r.",
				"- Use up/down to pick a stored request.",
				"- Press enter to send that request back to Build mode.",
			}, "\n")
		}

		return strings.Join([]string{
			"Forge notes",
			"",
			"- Use the top tabs to switch between Build Request and Load Collection.",
			"- Build mode lets you shape a request by hand.",
			"- Set Collection Path in Build mode to choose where ctrl+w writes the request.",
			"- Load mode lets you pull a stored request from a collection file.",
			"- Press ctrl+s to send from Build mode.",
			"- Press ctrl+w to save from Build mode.",
		}, "\n")
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "Status: %d\n", m.response.Status)
	fmt.Fprintf(&builder, "Duration: %s\n\n", m.response.Duration)
	builder.WriteString("Headers\n")
	builder.WriteString(formatHeaders(m.response.Headers))
	builder.WriteString("\n\nBody\n")
	builder.WriteString(string(m.response.Body))
	return builder.String()
}

func (m model) responseFocused() bool {
	if m.mode == buildMode {
		return m.focusIndex == buildFocusResponse
	}
	return m.focusIndex == loadFocusResponse
}

func (m model) footerText() string {
	if m.mode == loadMode {
		return "normal: h/j/k/l move • i insert • esc normal • ctrl+r load collection • enter temper into forge • ctrl+c quit"
	}
	return "normal: h/j/k/l move • i insert • esc normal • ctrl+s send • ctrl+w save to collection • ctrl+c quit"
}

func (m model) modeLabel() string {
	if m.insertMode {
		return "INSERT"
	}
	return "NORMAL"
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

func parseHeaders(raw string) (map[string][]string, error) {
	headers := make(map[string][]string)

	for lineNumber, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("header line %d must look like Key: Value", lineNumber+1)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, fmt.Errorf("header line %d has an empty key", lineNumber+1)
		}

		headers[key] = append(headers[key], value)
	}

	return headers, nil
}

func renderHeadersInput(headers map[string][]string) string {
	if len(headers) == 0 {
		return ""
	}

	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	lines := make([]string, 0)
	for _, key := range keys {
		for _, value := range headers[key] {
			lines = append(lines, fmt.Sprintf("%s: %s", key, value))
		}
	}

	return strings.Join(lines, "\n")
}

func formatHeaders(headers map[string][]string) string {
	if len(headers) == 0 {
		return "(none)"
	}

	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(headers))
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("%s: %s", key, strings.Join(headers[key], ", ")))
	}

	return strings.Join(lines, "\n")
}

func formatVariables(vars map[string]string) []string {
	keys := make([]string, 0, len(vars))
	for key := range vars {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("- %s = %s", key, vars[key]))
	}
	return lines
}

func findMethodIndex(method string) int {
	method = strings.ToUpper(strings.TrimSpace(method))
	for index, candidate := range methods {
		if candidate == method {
			return index
		}
	}
	return 0
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type styles struct {
	header       lipgloss.Style
	subheader    lipgloss.Style
	badgeNormal  lipgloss.Style
	badgeInsert  lipgloss.Style
	panel        lipgloss.Style
	sectionTitle lipgloss.Style
	label        lipgloss.Style
	footer       lipgloss.Style
	status       lipgloss.Style
	normalFocus  lipgloss.Style
	insertFocus  lipgloss.Style
	blurred      lipgloss.Style
	methodFocus  lipgloss.Style
	methodBlur   lipgloss.Style
	responseBox  lipgloss.Style
	responseHit  lipgloss.Style
	tabActive    lipgloss.Style
	tabInactive  lipgloss.Style
	modeHint     lipgloss.Style
	summaryBox   lipgloss.Style
	requestBox   lipgloss.Style
	requestHit   lipgloss.Style
}

func newStyles() styles {
	coal := lipgloss.Color("#000000")
	soot := lipgloss.Color("#000000")
	parchment := lipgloss.Color("#F3E7D0")
	brass := lipgloss.Color("#CFA15D")
	ember := lipgloss.Color("#D96C3D")
	steel := lipgloss.Color("#9AA3AD")
	ash := lipgloss.Color("#B9A88D")

	return styles{
		header:       lipgloss.NewStyle().Foreground(brass).Background(coal).Bold(true).Padding(0, 1),
		subheader:    lipgloss.NewStyle().Foreground(ash).MarginBottom(1),
		badgeNormal:  lipgloss.NewStyle().Foreground(coal).Background(steel).Bold(true).Padding(0, 1),
		badgeInsert:  lipgloss.NewStyle().Foreground(parchment).Background(ember).Bold(true).Padding(0, 1),
		panel:        lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(brass).Background(soot).Padding(1).MarginRight(1),
		sectionTitle: lipgloss.NewStyle().Foreground(parchment).Bold(true).MarginBottom(1),
		label:        lipgloss.NewStyle().Foreground(steel).Bold(true).MarginTop(1),
		footer:       lipgloss.NewStyle().Foreground(ash).MarginTop(1),
		status:       lipgloss.NewStyle().Foreground(ember).Bold(true).MarginBottom(1),
		normalFocus:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(brass).Foreground(parchment).Padding(0, 1),
		insertFocus:  lipgloss.NewStyle().Border(lipgloss.ThickBorder()).BorderForeground(ember).Foreground(parchment).Padding(0, 1),
		blurred:      lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(steel).Padding(0, 1),
		methodFocus:  lipgloss.NewStyle().Foreground(parchment).Background(ember).Bold(true).Padding(0, 2),
		methodBlur:   lipgloss.NewStyle().Foreground(parchment).Background(coal).Bold(true).Padding(0, 2),
		responseBox:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(steel).Padding(0, 1),
		responseHit:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ember).Padding(0, 1),
		tabActive:    lipgloss.NewStyle().Foreground(parchment).Background(ember).Bold(true).Padding(0, 2),
		tabInactive:  lipgloss.NewStyle().Foreground(steel).Background(coal).Padding(0, 2),
		modeHint:     lipgloss.NewStyle().Foreground(ash).MarginTop(1),
		summaryBox:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(steel).Padding(0, 1),
		requestBox:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(steel).Padding(0, 1),
		requestHit:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(ember).Padding(0, 1),
	}
}

func (s styles) inputWrap(focused bool, insertMode bool) lipgloss.Style {
	if focused {
		if insertMode {
			return s.insertFocus
		}
		return s.normalFocus
	}
	return s.blurred
}

func (s styles) methodField(focused bool) lipgloss.Style {
	if focused {
		return s.methodFocus
	}
	return s.methodBlur
}

func (s styles) responseWrap(focused bool) lipgloss.Style {
	if focused {
		return s.responseHit
	}
	return s.responseBox
}

func (s styles) requestListWrap(focused bool) lipgloss.Style {
	if focused {
		return s.requestHit
	}
	return s.requestBox
}

func (s styles) tab(active, focused bool) lipgloss.Style {
	style := s.tabInactive
	if active {
		style = s.tabActive
	}
	if focused {
		style = style.Copy().Underline(true)
	}
	return style
}

func (s styles) modeBadge(insertMode bool) lipgloss.Style {
	if insertMode {
		return s.badgeInsert
	}
	return s.badgeNormal
}
