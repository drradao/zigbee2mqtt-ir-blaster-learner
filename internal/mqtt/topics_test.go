package mqtt

import "testing"

func TestListenTopic(t *testing.T) {
	got := ListenTopic("mydevice")
	want := "zigbee2mqtt/mydevice"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestListenTopic_EmptyDevice(t *testing.T) {
	got := ListenTopic("")
	want := "zigbee2mqtt/"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPublishTopic(t *testing.T) {
	got := PublishTopic("mydevice")
	want := "zigbee2mqtt/mydevice/set"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestPublishTopic_EmptyDevice(t *testing.T) {
	got := PublishTopic("")
	want := "zigbee2mqtt//set"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
