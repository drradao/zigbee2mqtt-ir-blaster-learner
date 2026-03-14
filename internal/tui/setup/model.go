package setup

import (
	"fmt"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/drradao/zigbee2mqtt-ir-blaster-learner/internal/config"
	"github.com/drradao/zigbee2mqtt-ir-blaster-learner/internal/tui/styles"
)

// Mode controls which fields are shown in the setup form.
type Mode int

const (
	// ModeMissing shows only empty required fields (Host, Device).
	ModeMissing Mode = iota
	// ModeFull shows all fields — used by irlearn config.
	ModeFull
	// ModeLogin shows only User + Password — used by --login on ui.
	ModeLogin
)

// Result is returned by Run after the form exits.
type Result struct {
	Config  config.Config
	Saved   bool
	Aborted bool
	Err     error
}

// field indices within the fixed [5]Field array
const (
	idxHost     = 0
	idxPort     = 1
	idxDevice   = 2
	idxUser     = 3
	idxPassword = 4
)

// Model is the bubbletea model for the setup form.
type Model struct {
	fields     [5]Field
	visible    []int // ordered indices into fields shown for this Mode
	cursor     int   // index into visible
	mode       Mode
	configPath string
	atSave     bool // true when past the last field, awaiting y/n
	width      int
	height     int
	result     Result
	done       bool
}

// New creates a new Model pre-populated with an existing config.
func New(existing config.Config, mode Mode, configPath string) *Model {
	m := &Model{
		mode:       mode,
		configPath: configPath,
	}

	m.fields[idxHost] = NewField("Host", "192.168.1.100", false)
	m.fields[idxPort] = NewField("Port", "1883", false)
	m.fields[idxDevice] = NewField("Device", "living_room_ir", false)
	m.fields[idxUser] = NewField("User", "(optional)", false)
	m.fields[idxPassword] = NewField("Password", "(optional)", true)

	m.fields[idxHost].SetValue(existing.MQTT.Host)
	if existing.MQTT.Port != 0 {
		m.fields[idxPort].SetValue(fmt.Sprintf("%d", existing.MQTT.Port))
	}
	m.fields[idxDevice].SetValue(existing.Device)
	m.fields[idxUser].SetValue(existing.MQTT.User)
	m.fields[idxPassword].SetValue(existing.MQTT.Password)

	switch mode {
	case ModeMissing:
		for _, idx := range []int{idxHost, idxDevice} {
			if m.fields[idx].Value() == "" {
				m.visible = append(m.visible, idx)
			}
		}
	case ModeFull:
		m.visible = []int{idxHost, idxPort, idxDevice, idxUser, idxPassword}
	case ModeLogin:
		m.visible = []int{idxUser, idxPassword}
	}

	return m
}

// Result returns the form result after the model has exited.
func (m *Model) Result() Result { return m.result }

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if m.atSave {
			return m.handleSaveKey(msg)
		}
		if len(m.visible) == 0 {
			m.assembleResult()
			m.done = true
			return m, tea.Quit
		}

		field := &m.fields[m.visible[m.cursor]]
		advance, retreat, cancel := field.Update(msg)

		if cancel {
			m.result.Aborted = true
			m.done = true
			return m, tea.Quit
		}
		if advance {
			m.cursor++
			if m.cursor >= len(m.visible) {
				if m.mode == ModeLogin {
					m.assembleResult()
					m.done = true
					return m, tea.Quit
				}
				m.atSave = true
			}
		}
		if retreat && m.cursor > 0 {
			m.cursor--
		}
	}
	return m, nil
}

func (m *Model) handleSaveKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		m.assembleResult()
		if err := config.Save(m.configPath, &m.result.Config); err != nil {
			m.result.Err = err
		} else {
			m.result.Saved = true
		}
		m.done = true
		return m, tea.Quit
	case "n", "N", "enter", "esc":
		m.assembleResult()
		m.done = true
		return m, tea.Quit
	case "q", "ctrl+c":
		m.result.Aborted = true
		m.done = true
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) assembleResult() {
	port := 1883
	if portStr := m.fields[idxPort].Value(); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil && p >= 1 && p <= 65535 {
			port = p
		}
	}
	m.result.Config = config.Config{
		MQTT: config.MQTTConfig{
			Host:     m.fields[idxHost].Value(),
			Port:     port,
			User:     m.fields[idxUser].Value(),
			Password: m.fields[idxPassword].Value(),
		},
		Device: m.fields[idxDevice].Value(),
	}
}

// View implements tea.Model.
func (m *Model) View() string {
	if m.width == 0 {
		return ""
	}

	title := "irlearn setup"
	if m.mode == ModeLogin {
		title = "irlearn login"
	}

	// "Password" is 8 chars — the longest label
	labelWidth := 8

	var rows []string
	for i, idx := range m.visible {
		rows = append(rows, m.fields[idx].View(i == m.cursor && !m.atSave, labelWidth))
	}

	var lines []string
	lines = append(lines, styles.Title.Render(title))
	lines = append(lines, "")
	lines = append(lines, rows...)
	lines = append(lines, "")

	if m.atSave {
		lines = append(lines, styles.Prompt.Render("Save to config? [y/N] "))
	}

	lines = append(lines, styles.Muted.Render("Tab/Enter · Shift+Tab navigate · Esc abort"))

	form := lipgloss.JoinVertical(lipgloss.Left, lines...)
	formBox := styles.Border.Padding(1, 2).Render(form)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, formBox)
}
