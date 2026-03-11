package mqtt

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
)

// mqttOpTimeout is the maximum time to wait for MQTT operations to complete.
const mqttOpTimeout = 10 * time.Second

type pahoClient struct {
	opts     *pahomqtt.ClientOptions
	client   pahomqtt.Client
	logger   *slog.Logger
	mu       sync.Mutex
	handlers map[string]MessageHandler
	statusCh chan<- bool
}

func newPahoClient(host string, port int, user, password string, logger *slog.Logger) *pahoClient {
	p := &pahoClient{
		logger:   logger,
		handlers: make(map[string]MessageHandler),
	}

	opts := pahomqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", host, port))
	if user != "" {
		opts.SetUsername(user)
	}
	if password != "" {
		opts.SetPassword(password)
	}
	opts.SetClientID(fmt.Sprintf("irlearn-%d", time.Now().UnixNano()))
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetConnectionLostHandler(func(_ pahomqtt.Client, err error) {
		logger.Warn("MQTT connection lost", "error", err)
		p.sendStatus(false)
	})
	opts.SetOnConnectHandler(func(_ pahomqtt.Client) {
		logger.Info("MQTT connected")
		p.sendStatus(true)
	})

	p.opts = opts
	return p
}

func (p *pahoClient) Connect() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.client = pahomqtt.NewClient(p.opts)
	token := p.client.Connect()
	if !token.WaitTimeout(mqttOpTimeout) {
		return errors.New("MQTT connect timed out")
	}
	return token.Error()
}

func (p *pahoClient) Disconnect() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.client != nil && p.client.IsConnected() {
		p.client.Disconnect(250)
	}
}

func (p *pahoClient) Subscribe(topic string, handler MessageHandler) error {
	p.mu.Lock()
	p.handlers[topic] = handler
	p.mu.Unlock()

	token := p.client.Subscribe(topic, 1, func(_ pahomqtt.Client, msg pahomqtt.Message) {
		p.mu.Lock()
		h, ok := p.handlers[msg.Topic()]
		p.mu.Unlock()
		if ok {
			h(msg.Topic(), msg.Payload())
		}
	})
	if !token.WaitTimeout(mqttOpTimeout) {
		return errors.New("MQTT subscribe timed out")
	}
	return token.Error()
}

func (p *pahoClient) Publish(topic string, payload []byte) error {
	token := p.client.Publish(topic, 1, false, payload)
	if !token.WaitTimeout(mqttOpTimeout) {
		return errors.New("MQTT publish timed out")
	}
	return token.Error()
}

func (p *pahoClient) IsConnected() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.client != nil && p.client.IsConnected()
}

// SetStatusChannel registers the channel to receive connection state changes.
func (p *pahoClient) SetStatusChannel(ch chan<- bool) {
	p.mu.Lock()
	p.statusCh = ch
	p.mu.Unlock()
}

// sendStatus forwards a connection status change to the registered channel.
// The send is non-blocking to avoid stalling paho's internal goroutine.
func (p *pahoClient) sendStatus(connected bool) {
	p.mu.Lock()
	ch := p.statusCh
	p.mu.Unlock()
	if ch == nil {
		return
	}
	select {
	case ch <- connected:
	default:
	}
}
