package service_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/drradao/zigbee2mqtt-ir-blaster-learner/internal/service"
)

func TestReplay_Happy(t *testing.T) {
	client := &mockMQTTClient{}
	svc := service.NewReplayService(client, "zigbee2mqtt/mydevice/set")

	if err := svc.Replay("YWJjZGVm"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.publishCalls) != 1 {
		t.Fatalf("expected 1 Publish call, got %d", len(client.publishCalls))
	}
	call := client.publishCalls[0]
	if call.topic != "zigbee2mqtt/mydevice/set" {
		t.Errorf("topic: got %q", call.topic)
	}
	var m map[string]string
	if err := json.Unmarshal(call.payload, &m); err != nil {
		t.Fatalf("payload is not valid JSON: %v", err)
	}
	if m["ir_code_to_send"] != "YWJjZGVm" {
		t.Errorf("ir_code_to_send: got %q, want %q", m["ir_code_to_send"], "YWJjZGVm")
	}
}

func TestReplay_PublishError(t *testing.T) {
	sentinel := errors.New("broker down")
	client := &mockMQTTClient{publishErr: sentinel}
	svc := service.NewReplayService(client, "zigbee2mqtt/mydevice/set")

	err := svc.Replay("YWJj")
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestReplay_EmptyPayload(t *testing.T) {
	client := &mockMQTTClient{}
	svc := service.NewReplayService(client, "zigbee2mqtt/mydevice/set")

	if err := svc.Replay(""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(client.publishCalls) != 1 {
		t.Fatal("expected 1 Publish call")
	}
	var m map[string]string
	if err := json.Unmarshal(client.publishCalls[0].payload, &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["ir_code_to_send"] != "" {
		t.Errorf("expected empty string, got %q", m["ir_code_to_send"])
	}
}

func TestReplay_PayloadPreserved(t *testing.T) {
	client := &mockMQTTClient{}
	svc := service.NewReplayService(client, "zigbee2mqtt/mydevice/set")

	exactPayload := "BjAAAgAAAAAAAA=="
	if err := svc.Replay(exactPayload); err != nil {
		t.Fatal(err)
	}
	var m map[string]string
	if err := json.Unmarshal(client.publishCalls[0].payload, &m); err != nil {
		t.Fatal(err)
	}
	if m["ir_code_to_send"] != exactPayload {
		t.Errorf("payload not preserved: got %q, want %q", m["ir_code_to_send"], exactPayload)
	}
}
