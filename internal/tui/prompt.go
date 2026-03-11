package tui

import tea "github.com/charmbracelet/bubbletea"

// Prompt is a vim-style single-line text input shown at the bottom of the
// screen when the user needs to enter a name (e.g. when saving a command).
type Prompt struct {
	active bool
	value  string
	cursor int
}

// Active reports whether the prompt is currently visible.
func (p *Prompt) Active() bool { return p.active }

// Value returns the current text content of the prompt.
func (p *Prompt) Value() string { return p.value }

// Open clears and activates the prompt.
func (p *Prompt) Open() {
	p.active = true
	p.value = ""
	p.cursor = 0
}

// Close deactivates and clears the prompt.
func (p *Prompt) Close() {
	p.active = false
	p.value = ""
	p.cursor = 0
}

// Update handles a key message and returns whether the user submitted (Enter)
// or cancelled (Esc).
func (p *Prompt) Update(msg tea.Msg) (submit bool, cancel bool) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return false, false
	}
	switch km.Type {
	case tea.KeyEnter:
		return true, false
	case tea.KeyEsc:
		return false, true
	case tea.KeyBackspace, tea.KeyDelete:
		if p.cursor > 0 {
			runes := []rune(p.value)
			runes = append(runes[:p.cursor-1], runes[p.cursor:]...)
			p.value = string(runes)
			p.cursor--
		}
	case tea.KeyLeft:
		if p.cursor > 0 {
			p.cursor--
		}
	case tea.KeyRight:
		if p.cursor < len([]rune(p.value)) {
			p.cursor++
		}
	case tea.KeyRunes:
		runes := []rune(p.value)
		runes = append(runes[:p.cursor], append(km.Runes, runes[p.cursor:]...)...)
		p.value = string(runes)
		p.cursor += len(km.Runes)
	}
	return false, false
}

// View renders the prompt line. Returns an empty string when inactive.
func (p *Prompt) View() string {
	if !p.active {
		return ""
	}
	runes := []rune(p.value)
	before := string(runes[:p.cursor])
	after := string(runes[p.cursor:])
	return StylePrompt.Render(":") + " " + before + "█" + after
}
