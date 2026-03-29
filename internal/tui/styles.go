package tui

import "github.com/charmbracelet/lipgloss"

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
		style = style.Underline(true)
	}
	return style
}

func (s styles) modeBadge(insertMode bool) lipgloss.Style {
	if insertMode {
		return s.badgeInsert
	}
	return s.badgeNormal
}
