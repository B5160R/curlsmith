package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *model) renderForgePanel() string {
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		m.styles.sectionTitle.Render("Forge Bench"),
		m.renderModeTabs(),
	)

	if m.mode == buildMode {
		methodField := m.styles.methodField(m.focusIndex == int(buildFocusMethod)).Render(methods[m.methodIndex])
		nameField := m.styles.inputWrap(m.focusIndex == int(buildFocusName), m.insertMode).Render(m.nameInput.View())
		urlField := m.styles.inputWrap(m.focusIndex == int(buildFocusURL), m.insertMode).Render(m.urlInput.View())
		collectionPathField := m.styles.inputWrap(m.focusIndex == int(buildFocusCollectionPath), m.insertMode).Render(m.collectionPathInput.View())
		headersField := m.styles.inputWrap(m.focusIndex == int(buildFocusHeaders), m.insertMode).Render(m.headersInput.View())
		bodyField := m.styles.inputWrap(m.focusIndex == int(buildFocusBody), m.insertMode).Render(m.bodyInput.View())

		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			m.styles.modeHint.Render("Build mode shapes a request by hand. Press s to send it or w to write it into a collection file."),
			m.renderField("Method", methodField),
			m.renderField("Name", nameField),
			m.renderField("URL", urlField),
			m.renderField("Collection Path", collectionPathField),
			m.renderField("Headers", headersField),
			m.renderField("Body", bodyField),
		)
	} else {
		pathField := m.styles.inputWrap(m.focusIndex == int(loadFocusPath), m.insertMode).Render(m.collectionPathInput.View())
		requestField := m.styles.requestListWrap(m.focusIndex == int(loadFocusRequests)).Render(m.renderStoredRequests())

		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			m.styles.modeHint.Render("Load mode pulls a collection off the rack. Press r to read it, then enter on a request to temper it into the forge."),
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
			"Use the path field above and press r.",
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
				"- Enter a collection file path and press r.",
				"- Use up/down to pick a stored request.",
				"- Press enter to send that request back to Build mode.",
			}, "\n")
		}

		return strings.Join([]string{
			"Forge notes",
			"",
			"- Use the top tabs to switch between Build Request and Load Collection.",
			"- Build mode lets you shape a request by hand.",
			"- Set Collection Path in Build mode to choose where w writes the request.",
			"- Load mode lets you pull a stored request from a collection file.",
			"- Press s to send from Build mode.",
			"- Press w to save from Build mode.",
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
		return m.focusIndex == int(buildFocusResponse)
	}
	return m.focusIndex == int(loadFocusResponse)
}

func (m model) footerText() string {
	if m.mode == loadMode {
		return fmt.Sprintf("%s: h/j/k/l move • i insert • esc normal • r load collection • enter temper into forge • q quit", m.modeLabel())
	}
	return fmt.Sprintf("%s: h/j/k/l move • i insert • esc normal • s send • w save to collection • q quit", m.modeLabel())
}

func (m model) modeLabel() string {
	if m.insertMode {
		return "INSERT"
	}
	return "NORMAL"
}
