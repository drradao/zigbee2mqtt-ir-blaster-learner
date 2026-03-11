package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// View renders the complete TUI screen.
func (m *Model) View() string {
	if m.width == 0 {
		return "Loading…"
	}

	left := lipgloss.JoinVertical(lipgloss.Left,
		m.bufferPane.View(),
		m.sessionPane.View(),
	)
	right := m.previewPane.View()
	main := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	statusBar := m.renderStatusBar()

	promptLine := m.prompt.View()
	if promptLine == "" {
		promptLine = StyleMuted.Render("  [r]eplay  [s]ave  [n]ame  [d]elete  [C]lear  [l]earn  [L]ock  [Tab]switch  [q]uit")
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		main,
		statusBar,
		promptLine,
	)
}

func (m *Model) renderStatusBar() string {
	var connStatus string
	if m.connected {
		connStatus = StyleConnected.Render("MQTT: connected")
	} else {
		connStatus = StyleDisconnected.Render("MQTT: disconnected")
	}

	lockStr := ""
	if m.lockMode {
		lockStr = "  " + StyleConnected.Render("[LOCK]")
	}

	bufCount := fmt.Sprintf("  buf:%d", m.buf.Len())
	content := connStatus + lockStr + bufCount
	return StyleStatusBar.Width(m.width).Render(content)
}
