package panes

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/tui/styles"
)

// PreviewPane shows decoded details of the selected IR payload.
type PreviewPane struct {
	name       string
	payload    string // base64-encoded
	topic      string
	capturedAt time.Time
	width      int
	height     int
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

// SetMeta sets the MQTT topic and capture timestamp for the current selection.
func (p *PreviewPane) SetMeta(topic string, capturedAt time.Time) {
	p.topic = topic
	p.capturedAt = capturedAt
}

// wrapString splits s into lines of at most width runes.
func wrapString(s string, width int) []string {
	if width <= 0 || s == "" {
		return []string{s}
	}
	runes := []rune(s)
	var lines []string
	for len(runes) > width {
		lines = append(lines, string(runes[:width]))
		runes = runes[width:]
	}
	if len(runes) > 0 {
		lines = append(lines, string(runes))
	}
	return lines
}

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
		lines = append(lines, fmt.Sprintf("  Name:    %s", name))
		lines = append(lines, fmt.Sprintf("  Size:    %d bytes", len(decoded)))

		if p.topic != "" {
			display := strings.TrimPrefix(p.topic, "zigbee2mqtt/")
			lines = append(lines, fmt.Sprintf("  Topic:   %s", display))
		}

		if !p.capturedAt.IsZero() {
			now := time.Now()
			var ts string
			if p.capturedAt.Year() == now.Year() && p.capturedAt.YearDay() == now.YearDay() {
				ts = p.capturedAt.Format("15:04:05")
			} else {
				ts = p.capturedAt.Format("Jan 2 15:04")
			}
			lines = append(lines, fmt.Sprintf("  Captured: %s", ts))
		}

		// Payload section — word-wrapped base64.
		innerWidth := p.width - 6 // account for border + indent
		if innerWidth < 10 {
			innerWidth = 10
		}
		lines = append(lines, "")
		for _, chunk := range wrapString(p.payload, innerWidth) {
			lines = append(lines, "  "+chunk)
		}
	}

	content := title + "\n" + strings.Join(lines, "\n")
	return styles.Border.Width(p.width - 2).Height(p.height - 2).Render(content)
}
