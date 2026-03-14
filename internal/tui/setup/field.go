// Package setup implements a standalone bubbletea form used for first-run
// configuration and the irlearn config command. It must not import internal/tui
// to avoid import cycles.
package setup

import (
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/tui/styles"
)

// Field is a single text input field in the setup form.
type Field struct {
	Label       string
	Placeholder string
	masked      bool
	value       string
	cursor      int
}

// NewField creates a new Field with the given label, placeholder, and masked flag.
func NewField(label, placeholder string, masked bool) Field {
	return Field{
		Label:       label,
		Placeholder: placeholder,
		masked:      masked,
	}
}

// Value returns the current field value.
func (f *Field) Value() string { return f.value }

// SetValue pre-populates the field with an existing value.
func (f *Field) SetValue(v string) {
	f.value = v
	f.cursor = utf8.RuneCountInString(v)
}

// Update handles a key message. Returns navigation signals:
// advance moves to the next field, retreat moves to the previous, cancel aborts the form.
func (f *Field) Update(msg tea.KeyMsg) (advance, retreat, cancel bool) {
	switch msg.Type {
	case tea.KeyEnter, tea.KeyTab:
		return true, false, false
	case tea.KeyShiftTab:
		return false, true, false
	case tea.KeyEsc:
		return false, false, true
	case tea.KeyBackspace, tea.KeyDelete:
		if f.cursor > 0 {
			runes := []rune(f.value)
			runes = append(runes[:f.cursor-1], runes[f.cursor:]...)
			f.value = string(runes)
			f.cursor--
		}
	case tea.KeyLeft:
		if !f.masked && f.cursor > 0 {
			f.cursor--
		}
	case tea.KeyRight:
		if !f.masked {
			if f.cursor < utf8.RuneCountInString(f.value) {
				f.cursor++
			}
		}
	case tea.KeySpace:
		if f.masked {
			f.value += " "
			f.cursor = utf8.RuneCountInString(f.value)
		} else {
			runes := []rune(f.value)
			tail := make([]rune, len(runes)-f.cursor)
			copy(tail, runes[f.cursor:])
			runes = append(runes[:f.cursor], append([]rune{' '}, tail...)...)
			f.value = string(runes)
			f.cursor++
		}
	case tea.KeyRunes:
		if f.masked {
			f.value += string(msg.Runes)
			f.cursor = utf8.RuneCountInString(f.value)
		} else {
			runes := []rune(f.value)
			tail := make([]rune, len(runes)-f.cursor)
			copy(tail, runes[f.cursor:])
			runes = append(runes[:f.cursor], append(msg.Runes, tail...)...)
			f.value = string(runes)
			f.cursor += len(msg.Runes)
		}
	}
	return false, false, false
}

// View renders the field. labelWidth is used to right-pad the label for alignment.
func (f *Field) View(active bool, labelWidth int) string {
	label := f.Label
	if len(label) < labelWidth {
		label += strings.Repeat(" ", labelWidth-len(label))
	}

	border := styles.Border
	if active {
		border = styles.ActiveBorder
	}

	var input string
	if f.value == "" && !active {
		input = styles.Muted.Render(f.Placeholder)
	} else if f.masked {
		dots := strings.Repeat("•", utf8.RuneCountInString(f.value))
		if active {
			input = dots + "█"
		} else {
			input = dots
		}
	} else {
		runes := []rune(f.value)
		before := string(runes[:f.cursor])
		after := string(runes[f.cursor:])
		if active {
			input = before + "█" + after
		} else {
			input = f.value
		}
	}

	inputBox := border.Padding(0, 1).Render(input)
	return lipgloss.JoinHorizontal(lipgloss.Center, styles.Muted.Render(label+" "), inputBox)
}
