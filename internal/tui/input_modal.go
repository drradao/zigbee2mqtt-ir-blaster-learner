package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// inputModal is a centered single-field modal for collecting a text name.
// It is used for both "save command" and "rename command" flows.
type inputModal struct {
	active  bool
	title   string
	context string // "save" or "rename"
	field   Prompt
}

// Open activates the modal with the given title and context tag.
func (m *inputModal) Open(title, context string) {
	m.active = true
	m.title = title
	m.context = context
	m.field.Open()
}

// OpenWithValue activates the modal with a pre-populated field value.
// The cursor is placed at the end of the value.
func (m *inputModal) OpenWithValue(title, context, initial string) {
	m.active = true
	m.title = title
	m.context = context
	m.field.Open()
	m.field.value = initial
	m.field.cursor = len([]rune(initial))
}

// IsActive reports whether the modal is currently shown.
func (m *inputModal) IsActive() bool { return m.active }

// Context returns the context string set when opening.
func (m *inputModal) Context() string { return m.context }

// Update handles key input. Returns submitted/dismissed signals and the entered value.
func (m *inputModal) Update(msg tea.KeyMsg) (submitted, dismissed bool, value string) {
	submit, cancel := m.field.Update(msg)
	if cancel {
		m.active = false
		return false, true, ""
	}
	if submit {
		v := m.field.value
		m.active = false
		return true, false, v
	}
	return false, false, ""
}

// View renders the centered input modal box over the full terminal.
func (m *inputModal) View(width, height int) string {
	fieldDisplay := renderModalField(m.field.value, m.field.cursor, false, true)
	inputLine := lipgloss.JoinHorizontal(lipgloss.Center,
		StyleMuted.Render("Name "), fieldDisplay)

	footer := StyleMuted.Render("Enter confirm · Esc dismiss")

	form := lipgloss.JoinVertical(lipgloss.Left,
		StyleTitle.Render(m.title),
		"",
		inputLine,
		"",
		footer,
	)
	box := StyleActiveBorder.Padding(1, 2).Render(form)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
