package panes

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/session"
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/tui/styles"
)

// SessionPane displays the commands saved in the current session file.
type SessionPane struct {
	sess     *session.Session
	cursor   int
	width    int
	height   int
	active   bool
	filename string
}

// NewSessionPane creates a SessionPane labelled with the given filename.
func NewSessionPane(filename string) *SessionPane {
	return &SessionPane{filename: filename}
}

// SetActive marks this pane as focused.
func (p *SessionPane) SetActive(a bool) { p.active = a }

// SetSize updates render dimensions.
func (p *SessionPane) SetSize(w, h int) { p.width = w; p.height = h }

// SetSession replaces the displayed session and keeps cursor in bounds.
func (p *SessionPane) SetSession(s *session.Session) {
	p.sess = s
	if p.sess != nil && p.cursor >= len(p.sess.Commands) && len(p.sess.Commands) > 0 {
		p.cursor = len(p.sess.Commands) - 1
	}
}

// Cursor returns the current row index.
func (p *SessionPane) Cursor() int { return p.cursor }

// Selected returns a copy of the currently highlighted command, or nil.
func (p *SessionPane) Selected() *session.Command {
	if p.sess == nil || len(p.sess.Commands) == 0 {
		return nil
	}
	if p.cursor < 0 || p.cursor >= len(p.sess.Commands) {
		return nil
	}
	cmd := p.sess.Commands[p.cursor]
	return &cmd
}

// Update handles keyboard navigation within this pane.
func (p *SessionPane) Update(msg tea.Msg) {
	if p.sess == nil {
		return
	}
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
		if p.cursor < len(p.sess.Commands)-1 {
			p.cursor++
		}
	}
}

// View renders the pane.
func (p *SessionPane) View() string {
	name := p.filename
	if name == "" {
		name = "session.yaml"
	}
	title := styles.Title.Render("Session: " + name)

	var rows []string
	if p.sess != nil {
		for i, cmd := range p.sess.Commands {
			if i == p.cursor {
				rows = append(rows, styles.Selected.Width(p.width-4).Render("> "+cmd.Name))
			} else {
				rows = append(rows, "  "+cmd.Name)
			}
		}
	}
	if len(rows) == 0 {
		rows = append(rows, styles.Muted.Render("  (no commands)"))
	}

	content := title + "\n" + strings.Join(rows, "\n")
	border := styles.Border
	if p.active {
		border = styles.ActiveBorder
	}
	return border.Width(p.width - 2).Height(p.height - 2).Render(content)
}
