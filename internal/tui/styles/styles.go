// Package styles contains the lipgloss style definitions shared across the
// tui and tui/panes packages.
package styles

import "github.com/charmbracelet/lipgloss"

var (
	ColorPrimary  = lipgloss.Color("#7C3AED")
	ColorAccent   = lipgloss.Color("#10B981")
	ColorMuted    = lipgloss.Color("#6B7280")
	ColorError    = lipgloss.Color("#EF4444")
	ColorSelected = lipgloss.Color("#1D4ED8")

	// Border is the default pane border (inactive).
	Border = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorMuted)

	// ActiveBorder is the border for the currently focused pane.
	ActiveBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary)

	// Selected highlights the cursor row in a list.
	Selected = lipgloss.NewStyle().
			Background(ColorSelected).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	// Normal is the default unstyled text style.
	Normal = lipgloss.NewStyle()

	// StatusBar is the bottom status line background.
	StatusBar = lipgloss.NewStyle().
			Background(lipgloss.Color("#1F2937")).
			Foreground(lipgloss.Color("#F9FAFB")).
			Padding(0, 1)

	// Connected renders the "connected" indicator in green.
	Connected = lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true)

	// Disconnected renders the "disconnected" indicator in red.
	Disconnected = lipgloss.NewStyle().
			Foreground(ColorError).
			Bold(true)

	// Prompt renders the vim-style ":" prompt prefix.
	Prompt = lipgloss.NewStyle().
		Foreground(ColorPrimary)

	// Muted renders de-emphasised helper text.
	Muted = lipgloss.NewStyle().
		Foreground(ColorMuted)

	// Title renders pane title text.
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary)
)
