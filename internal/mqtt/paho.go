package mqtt

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
)

type pahoClient struct {
	opts     *pahomqtt.ClientOptions
	client   pahomqtt.Client
	logger   *slog.Logger
	mu       sync.Mutex
	handlers map[string]MessageHandler
}

func newPahoClient(host string, port int, user, password string, logger *slog.Logger) *pahoClient {
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
	})
	opts.SetOnConnectHandler(func(_ pahomqtt.Client) {
		logger.Info("MQTT connected")
	})

	return &pahoClient{
		opts:     opts,
		logger:   logger,
		handlers: make(map[string]MessageHandler),
	}
}

func (p *pahoClient) Connect() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.client = pahomqtt.NewClient(p.opts)
	token := p.client.Connect()
	token.Wait()
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
	token.Wait()
	return token.Error()
}

func (p *pahoClient) Publish(topic string, payload []byte) error {
	token := p.client.Publish(topic, 1, false, payload)
	token.Wait()
	return token.Error()
}

func (p *pahoClient) IsConnected() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.client != nil && p.client.IsConnected()
}
