package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// loginModal is an inline credentials form overlaid on the full screen.
// It is shown when --login is active inside the TUI, acting as a forward-
// compatible hook for future mid-session credential re-entry (e.g. reconnect).
// In the current flow, actual credential collection happens in pre-flight
// (cmd/preflight.go) before MQTT connects; this modal dispatches LoginSubmitted.
type loginModal struct {
	active    bool
	userField Prompt
	passField Prompt
	onPass    bool // false = user field focused, true = password field focused
}

func newLoginModal() loginModal {
	lm := loginModal{}
	lm.passField.Masked = true
	return lm
}

// Open activates the modal and resets both fields.
func (l *loginModal) Open() {
	l.active = true
	l.onPass = false
	l.userField.Open()
	l.passField.Open()
}

// IsActive reports whether the modal is currently shown.
func (l *loginModal) IsActive() bool { return l.active }

// Update handles key input. Returns submitted/dismissed signals and credentials.
func (l *loginModal) Update(msg tea.KeyMsg) (submitted, dismissed bool, user, pass string) {
	if !l.onPass {
		submit, cancel := l.userField.Update(msg)
		if cancel {
			l.active = false
			return false, true, "", ""
		}
		if submit {
			l.onPass = true
		}
		return false, false, "", ""
	}

	submit, cancel := l.passField.Update(msg)
	if cancel {
		l.active = false
		return false, true, "", ""
	}
	if submit {
		u := l.userField.value
		p := l.passField.value
		l.active = false
		return true, false, u, p
	}
	return false, false, "", ""
}

// View renders the centered login modal box over the full terminal.
func (l *loginModal) View(width, height int) string {
	userDisplay := renderModalField(l.userField.value, l.userField.cursor, false, !l.onPass)
	passDisplay := renderModalField(l.passField.value, l.passField.cursor, true, l.onPass)

	userLine := lipgloss.JoinHorizontal(lipgloss.Center,
		StyleMuted.Render("User     "), userDisplay)
	passLine := lipgloss.JoinHorizontal(lipgloss.Center,
		StyleMuted.Render("Password "), passDisplay)

	footer := StyleMuted.Render("Enter confirm · Esc dismiss")

	form := lipgloss.JoinVertical(lipgloss.Left,
		StyleTitle.Render("irlearn login"),
		"",
		userLine,
		passLine,
		"",
		footer,
	)
	box := StyleActiveBorder.Padding(1, 2).Render(form)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}

// renderModalField renders a single field inside the login modal.
func renderModalField(value string, cursor int, masked, active bool) string {
	border := StyleBorder
	if active {
		border = StyleActiveBorder
	}

	var content string
	if masked {
		dots := strings.Repeat("•", len([]rune(value)))
		if active {
			content = dots + "█"
		} else {
			content = dots
		}
	} else {
		runes := []rune(value)
		before := string(runes[:cursor])
		after := string(runes[cursor:])
		if active {
			content = before + "█" + after
		} else {
			content = value
		}
	}

	return border.Padding(0, 1).Render(content)
}
