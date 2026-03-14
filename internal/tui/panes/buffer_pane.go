// Package panes contains the individual TUI pane components.
package panes

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/buffer"
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/tui/styles"
)

// BufferPane displays the in-memory queue of received IR messages.
type BufferPane struct {
	msgs   []buffer.IRMessage
	cursor int
	width  int
	height int
	active bool
}

// NewBufferPane creates an empty BufferPane.
func NewBufferPane() *BufferPane {
	return &BufferPane{}
}

// SetActive marks this pane as focused so it renders with the active border.
func (p *BufferPane) SetActive(a bool) { p.active = a }

// SetSize updates the render dimensions.
func (p *BufferPane) SetSize(w, h int) { p.width = w; p.height = h }

// SetMessages replaces the displayed message list and keeps cursor in bounds.
func (p *BufferPane) SetMessages(msgs []buffer.IRMessage) {
	p.msgs = msgs
	if p.cursor >= len(p.msgs) && len(p.msgs) > 0 {
		p.cursor = len(p.msgs) - 1
	}
}

// Cursor returns the current row index.
func (p *BufferPane) Cursor() int { return p.cursor }

// Selected returns a copy of the currently highlighted message, or nil.
func (p *BufferPane) Selected() *buffer.IRMessage {
	if len(p.msgs) == 0 || p.cursor < 0 || p.cursor >= len(p.msgs) {
		return nil
	}
	m := p.msgs[p.cursor]
	return &m
}

// Update handles keyboard navigation within this pane.
func (p *BufferPane) Update(msg tea.Msg) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return
	}
	switch km.String() {
	case "up", "k":
		if p.cursor > 0 {
			p.cursor--
		}
	case "down", "j":
		if p.cursor < len(p.msgs)-1 {
			p.cursor++
		}
	}
}

// View renders the pane.
func (p *BufferPane) View() string {
	title := styles.Title.Render("Buffer")
	var rows []string
	for i, msg := range p.msgs {
		label := fmt.Sprintf("msg %d  (%d bytes)", i+1, len(msg.Payload))
		if i == p.cursor {
			rows = append(rows, styles.Selected.Width(p.width-4).Render("> "+label))
		} else {
			rows = append(rows, "  "+label)
		}
	}
	if len(rows) == 0 {
		rows = append(rows, styles.Muted.Render("  (no messages)"))
	}
	content := title + "\n" + strings.Join(rows, "\n")

	border := styles.Border
	if p.active {
		border = styles.ActiveBorder
	}
	return border.Width(p.width - 2).Height(p.height - 2).Render(content)
}
