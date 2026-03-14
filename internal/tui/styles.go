// styles.go re-exports the shared style definitions from the styles sub-package
// so that code within the tui package can access them without a qualifier.
package tui

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/drradao/zigbee2mqtt-ir-blaster-learner/internal/tui/styles"
)

// Re-export shared styles under the names used by tui-package code.
var (
	StyleBorder       = styles.Border
	StyleActiveBorder = styles.ActiveBorder
	StyleSelected     = styles.Selected
	StyleNormal       = styles.Normal
	StyleStatusBar    = styles.StatusBar
	StyleConnected    = styles.Connected
	StyleDisconnected = styles.Disconnected
	StylePrompt       = styles.Prompt
	StyleMuted        = styles.Muted
	StyleTitle        = styles.Title
)

// Ensure lipgloss is used (it is referenced by styles sub-package).
var _ = lipgloss.NewStyle()
