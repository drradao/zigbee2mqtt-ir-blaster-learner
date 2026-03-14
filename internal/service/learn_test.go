package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/drradao/zigbee2mqtt-ir-blaster-learner/internal/service"
)

func TestSendLearnCommand_Happy(t *testing.T) {
	client := &mockMQTTClient{}
	svc := service.NewLearnService(client, "zigbee2mqtt/mydevice/set")

	if err := svc.SendLearnCommand(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.publishCalls) != 1 {
		t.Fatalf("expected 1 Publish call, got %d", len(client.publishCalls))
	}
	call := client.publishCalls[0]
	if call.topic != "zigbee2mqtt/mydevice/set" {
		t.Errorf("topic: got %q", call.topic)
	}
	if string(call.payload) != `{"learn_ir_code":true}` {
		t.Errorf("payload: got %q", call.payload)
	}
}

func TestSendLearnCommand_PublishError(t *testing.T) {
	sentinel := errors.New("broker unreachable")
	client := &mockMQTTClient{publishErr: sentinel}
	svc := service.NewLearnService(client, "zigbee2mqtt/mydevice/set")

	err := svc.SendLearnCommand()
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestLastSent_ZeroBeforeFirstCall(t *testing.T) {
	client := &mockMQTTClient{}
	svc := service.NewLearnService(client, "zigbee2mqtt/mydevice/set")

	if !svc.LastSent().IsZero() {
		t.Errorf("expected zero time before any send, got %v", svc.LastSent())
	}
}

func TestLastSent_UpdatedAfterSend(t *testing.T) {
	client := &mockMQTTClient{}
	svc := service.NewLearnService(client, "zigbee2mqtt/mydevice/set")

	before := time.Now()
	if err := svc.SendLearnCommand(); err != nil {
		t.Fatal(err)
	}
	after := time.Now()

	last := svc.LastSent()
	if last.Before(before) {
		t.Errorf("LastSent %v is before start of test %v", last, before)
	}
	if last.After(after) {
		t.Errorf("LastSent %v is after end of test %v", last, after)
	}
}
