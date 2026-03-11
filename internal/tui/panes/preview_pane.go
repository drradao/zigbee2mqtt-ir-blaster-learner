package panes

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/tui/styles"
)

// PreviewPane shows decoded details of the selected IR payload.
type PreviewPane struct {
	name    string
	payload string // base64-encoded
	width   int
	height  int
}

// NewPreviewPane creates an empty PreviewPane.
func NewPreviewPane() *PreviewPane {
	return &PreviewPane{}
}

// SetSize updates render dimensions.
func (p *PreviewPane) SetSize(w, h int) { p.width = w; p.height = h }

// SetName sets the display name for the currently previewed item.
func (p *PreviewPane) SetName(name string) { p.name = name }

// SetPayload updates the base64-encoded payload to preview.
func (p *PreviewPane) SetPayload(b64 string) { p.payload = b64 }

// View renders the pane.
func (p *PreviewPane) View() string {
	title := styles.Title.Render("Preview")
	var lines []string

	decoded, err := base64.StdEncoding.DecodeString(p.payload)
	if err != nil || len(decoded) == 0 {
		lines = append(lines, styles.Muted.Render("  No selection"))
	} else {
		name := p.name
		if name == "" {
			name = "—"
		}
		lines = append(lines, fmt.Sprintf("  Name:  %s", name))
		lines = append(lines, fmt.Sprintf("  Bytes: %d", len(decoded)))

		limit := 16
		if len(decoded) < limit {
			limit = len(decoded)
		}
		hexParts := make([]string, limit)
		for i, b := range decoded[:limit] {
			hexParts[i] = fmt.Sprintf("%02X", b)
		}
		hexStr := strings.Join(hexParts, " ")
		if len(decoded) > 16 {
			hexStr += " …"
		}
		lines = append(lines, fmt.Sprintf("  Hex:   %s", hexStr))
	}

	content := title + "\n" + strings.Join(lines, "\n")
	return styles.Border.Width(p.width - 2).Height(p.height - 2).Render(content)
}
