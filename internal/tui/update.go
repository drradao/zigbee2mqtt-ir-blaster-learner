package tui

import (
	"encoding/json"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/session"
)

// lockDebounce is the minimum interval between auto learn sends in lock mode.
const lockDebounce = 200 * time.Millisecond

// Update is the bubbletea update function.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Login modal captures all key events when active.
	if m.loginModal.IsActive() {
		if km, ok := msg.(tea.KeyMsg); ok {
			submitted, dismissed, user, pass := m.loginModal.Update(km)
			if submitted {
				return m, func() tea.Msg { return LoginSubmitted{User: user, Password: pass} }
			}
			if dismissed {
				return m, func() tea.Msg { return LoginDismissed{} }
			}
			return m, nil
		}
	}

	// Route key events to the input modal when it is active.
	if m.inputModal.IsActive() {
		if km, ok := msg.(tea.KeyMsg); ok {
			submitted, dismissed, val := m.inputModal.Update(km)
			if dismissed {
				return m, nil
			}
			if submitted {
				return m, m.handleInputSubmit(val)
			}
			return m, nil
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layoutPanes()
		return m, nil

	case MQTTMessageReceived:
		m.buf.Add(msg.Msg)
		m.bufferPane.SetMessages(m.buf.All())
		m.updatePreview()
		cmd := waitForMQTT(m.mqttCh)
		if m.lockMode && time.Since(m.learnSvc.LastSent()) >= lockDebounce {
			_ = m.learnSvc.SendLearnCommand()
		}
		return m, cmd

	case StatusChanged:
		m.connected = msg.Connected
		return m, waitForStatus(m.statusCh)

	case LoginSubmitted:
		// Credentials collected from the modal — modal is already closed.
		// The in-TUI modal is a forward hook; credentials from pre-flight
		// are already applied to the active MQTT connection.
		return m, nil

	case LoginDismissed:
		// Modal dismissed without submitting.
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "tab":
		if m.active == paneBuffer {
			m.active = paneSession
		} else {
			m.active = paneBuffer
		}
		m.bufferPane.SetActive(m.active == paneBuffer)
		m.sessionPane.SetActive(m.active == paneSession)
		m.updatePreview()

	case "up", "k", "down", "j":
		if m.active == paneBuffer {
			m.bufferPane.Update(msg)
		} else {
			m.sessionPane.Update(msg)
		}
		m.updatePreview()

	case "r":
		return m, m.replaySelected()

	case "s":
		if m.active == paneBuffer && m.bufferPane.Selected() != nil {
			m.inputModal.Open("Save command", "save")
		}

	case "n":
		if m.active == paneSession {
			idx := m.sessionPane.Cursor()
			if m.sess != nil && idx >= 0 && idx < len(m.sess.Commands) {
				current := m.sess.Commands[idx].Name
				m.inputModal.OpenWithValue("Rename command", "rename", current)
			}
		} else {
			m.inputModal.Open("Rename command", "rename")
		}

	case "d":
		m.deleteSelected()

	case "C":
		m.buf.Clear()
		m.bufferPane.SetMessages(nil)
		m.updatePreview()

	case "l":
		_ = m.learnSvc.SendLearnCommand()

	case "L":
		m.lockMode = !m.lockMode

	case "y":
		m.yankPayload()

	case "Y":
		m.yankSendCommand()
	}
	return m, nil
}

func (m *Model) handleInputSubmit(val string) tea.Cmd {
	switch m.inputModal.Context() {
	case "save":
		sel := m.bufferPane.Selected()
		if sel != nil && val != "" {
			cmd := session.Command{
				Name:       val,
				Payload:    string(sel.Payload),
				CapturedAt: sel.ReceivedAt,
			}
			m.sess.Commands = append(m.sess.Commands, cmd)
			m.sessionPane.SetSession(m.sess)
			_ = m.sessionRepo.Save(m.sess)
		}

	case "rename":
		if m.active == paneSession && val != "" {
			idx := m.sessionPane.Cursor()
			if m.sess != nil && idx >= 0 && idx < len(m.sess.Commands) {
				m.sess.Commands[idx].Name = val
				m.sessionPane.SetSession(m.sess)
				_ = m.sessionRepo.Save(m.sess)
			}
		}
	}
	return nil
}

func (m *Model) yankPayload() {
	b64 := m.selectedPayload()
	if b64 != "" {
		_ = clipboard.WriteAll(b64)
	}
}

func (m *Model) yankSendCommand() {
	b64 := m.selectedPayload()
	if b64 != "" {
		js, _ := json.Marshal(map[string]string{"ir_code_to_send": b64})
		_ = clipboard.WriteAll(string(js))
	}
}

// selectedPayload returns the base64 payload for whichever pane is active.
func (m *Model) selectedPayload() string {
	if m.active == paneBuffer {
		if sel := m.bufferPane.Selected(); sel != nil {
			return string(sel.Payload)
		}
	} else {
		if sel := m.sessionPane.Selected(); sel != nil {
			return sel.Payload
		}
	}
	return ""
}

func (m *Model) replaySelected() tea.Cmd {
	if m.active == paneBuffer {
		sel := m.bufferPane.Selected()
		if sel != nil {
			b64 := string(sel.Payload)
			_ = m.replaySvc.Replay(b64)
		}
	} else {
		sel := m.sessionPane.Selected()
		if sel != nil {
			_ = m.replaySvc.Replay(sel.Payload)
		}
	}
	return nil
}

func (m *Model) deleteSelected() {
	if m.active == paneBuffer {
		idx := m.bufferPane.Cursor()
		m.buf.Remove(idx)
		m.bufferPane.SetMessages(m.buf.All())
	} else {
		idx := m.sessionPane.Cursor()
		if m.sess != nil && idx >= 0 && idx < len(m.sess.Commands) {
			m.sess.Commands = append(m.sess.Commands[:idx], m.sess.Commands[idx+1:]...)
			m.sessionPane.SetSession(m.sess)
			_ = m.sessionRepo.Save(m.sess)
		}
	}
	m.updatePreview()
}

func (m *Model) updatePreview() {
	if m.active == paneBuffer {
		sel := m.bufferPane.Selected()
		if sel != nil {
			m.previewPane.SetName("")
			m.previewPane.SetPayload(string(sel.Payload))
			m.previewPane.SetMeta(sel.Topic, sel.ReceivedAt)
		} else {
			m.previewPane.SetName("")
			m.previewPane.SetPayload("")
			m.previewPane.SetMeta("", time.Time{})
		}
	} else {
		sel := m.sessionPane.Selected()
		if sel != nil {
			m.previewPane.SetName(sel.Name)
			m.previewPane.SetPayload(sel.Payload)
			m.previewPane.SetMeta("", sel.CapturedAt)
		} else {
			m.previewPane.SetName("")
			m.previewPane.SetPayload("")
			m.previewPane.SetMeta("", time.Time{})
		}
	}
}

func (m *Model) layoutPanes() {
	if m.width == 0 || m.height == 0 {
		return
	}
	// Reserve 2 lines at the bottom: status bar + prompt line.
	const bottomReserve = 2
	contentH := m.height - bottomReserve

	leftW := m.width * 6 / 10
	rightW := m.width - leftW

	// Split the left column between buffer (top) and session (bottom).
	bufH := contentH / 2
	sessH := contentH - bufH

	m.bufferPane.SetSize(leftW, bufH)
	m.sessionPane.SetSize(leftW, sessH)
	m.previewPane.SetSize(rightW, contentH)
}
