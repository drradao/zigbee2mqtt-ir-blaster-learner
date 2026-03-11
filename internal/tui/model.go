package tui

import (
	"encoding/json"
	"log/slog"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/buffer"
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/mqtt"
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/service"
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/session"
	"github.com/dadao/zigbee2mqtt-ir-blaster-learner/internal/tui/panes"
)

type activePane int

const (
	paneBuffer activePane = iota
	paneSession
)

// Model is the root bubbletea model for the application.
type Model struct {
	buf         *buffer.IRBuffer
	mqttClient  mqtt.MQTTClient
	learnSvc    *service.LearnService
	replaySvc   *service.ReplayService
	sessionRepo session.SessionRepository
	sess        *session.Session
	// mqttCh is written to by the MQTT message handler goroutine and read by
	// the bubbletea waitForMQTT command to deliver messages into the event loop.
	mqttCh chan buffer.IRMessage
	// statusCh is written to by the paho status callbacks (true=connected,
	// false=lost) and read by waitForStatus to update the status bar.
	statusCh    <-chan bool
	logger      *slog.Logger
	sessionFile string

	bufferPane  *panes.BufferPane
	sessionPane *panes.SessionPane
	previewPane *panes.PreviewPane
	inputModal  inputModal
	loginModal  loginModal

	active    activePane
	connected bool
	lockMode  bool
	width     int
	height    int
}

// ModelParams groups all dependencies required to construct a Model.
type ModelParams struct {
	Buf         *buffer.IRBuffer
	MQTTClient  mqtt.MQTTClient
	LearnSvc    *service.LearnService
	ReplaySvc   *service.ReplayService
	SessionRepo session.SessionRepository
	Logger      *slog.Logger
	SessionFile string
	MQTTCh      chan buffer.IRMessage
	StatusCh    <-chan bool
	// LoginMode, when true, opens the credentials modal immediately on start.
	// Pre-flight (cmd/preflight.go) handles the actual credential injection
	// before MQTT connects; the modal here is a forward-compatible hook for
	// future mid-session credential re-entry.
	LoginMode bool
}

// NewModel constructs the root model and loads the session from disk.
func NewModel(p ModelParams) *Model {
	sess, _ := p.SessionRepo.Load()
	if sess == nil {
		sess = &session.Session{}
	}
	m := &Model{
		buf:         p.Buf,
		mqttClient:  p.MQTTClient,
		learnSvc:    p.LearnSvc,
		replaySvc:   p.ReplaySvc,
		sessionRepo: p.SessionRepo,
		sess:        sess,
		mqttCh:      p.MQTTCh,
		statusCh:    p.StatusCh,
		connected:   p.MQTTClient.IsConnected(),
		logger:      p.Logger,
		sessionFile: p.SessionFile,
		bufferPane:  panes.NewBufferPane(),
		sessionPane: panes.NewSessionPane(p.SessionFile),
		previewPane: panes.NewPreviewPane(),
		active:      paneBuffer,
		loginModal:  newLoginModal(),
	}
	m.bufferPane.SetActive(true)
	m.sessionPane.SetSession(sess)
	if p.LoginMode {
		m.loginModal.Open()
	}
	return m
}

// Init starts the MQTT listener and status watcher commands.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(waitForMQTT(m.mqttCh), waitForStatus(m.statusCh))
}

// waitForMQTT blocks until a message arrives on the channel, then returns it
// as a bubbletea message so it lands in the Update loop on the main goroutine.
func waitForMQTT(ch chan buffer.IRMessage) tea.Cmd {
	return func() tea.Msg {
		return MQTTMessageReceived{Msg: <-ch}
	}
}

// waitForStatus blocks until a connection status change arrives on the channel.
func waitForStatus(ch <-chan bool) tea.Cmd {
	return func() tea.Msg {
		return StatusChanged{Connected: <-ch}
	}
}

// extractLearnedCode pulls the learned_ir_code string from a raw MQTT payload.
func extractLearnedCode(payload []byte) string {
	var m map[string]interface{}
	if err := json.Unmarshal(payload, &m); err != nil {
		return ""
	}
	v, ok := m["learned_ir_code"]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}
