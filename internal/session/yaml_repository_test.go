package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_MissingFile(t *testing.T) {
	r := newYAMLRepository(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	s, err := r.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil Session")
	}
	if len(s.Commands) != 0 {
		t.Errorf("expected empty Commands, got %d", len(s.Commands))
	}
}

func TestLoad_ValidYAML(t *testing.T) {
	content := `
commands:
  - name: power
    payload: YWJj
    captured_at: 2024-01-01T00:00:00Z
`
	path := filepath.Join(t.TempDir(), "session.yaml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	r := newYAMLRepository(path)
	s, err := r.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(s.Commands))
	}
	if s.Commands[0].Name != "power" {
		t.Errorf("name: got %q", s.Commands[0].Name)
	}
	if s.Commands[0].Payload != "YWJj" {
		t.Errorf("payload: got %q", s.Commands[0].Payload)
	}
}

func TestLoad_MalformedYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.yaml")
	if err := os.WriteFile(path, []byte(":\t:bad{{"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := newYAMLRepository(path)
	_, err := r.Load()
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}
}

func TestSave_CreatesDirectories(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a", "b", "session.yaml")
	r := newYAMLRepository(path)
	if err := r.Save(&Session{}); err != nil {
		t.Fatalf("Save error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}

func TestSave_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.yaml")
	r := newYAMLRepository(path)
	ts := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	original := &Session{
		Commands: []Command{
			{Name: "vol-up", Payload: "cGF5bG9hZA==", CapturedAt: ts},
			{Name: "vol-down", Payload: "b3RoZXI=", CapturedAt: ts},
		},
	}
	if err := r.Save(original); err != nil {
		t.Fatal(err)
	}
	loaded, err := r.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(loaded.Commands))
	}
	for i, cmd := range original.Commands {
		if loaded.Commands[i].Name != cmd.Name {
			t.Errorf("[%d] name: got %q, want %q", i, loaded.Commands[i].Name, cmd.Name)
		}
		if loaded.Commands[i].Payload != cmd.Payload {
			t.Errorf("[%d] payload: got %q, want %q", i, loaded.Commands[i].Payload, cmd.Payload)
		}
		if !loaded.Commands[i].CapturedAt.Equal(cmd.CapturedAt) {
			t.Errorf("[%d] captured_at: got %v, want %v", i, loaded.Commands[i].CapturedAt, cmd.CapturedAt)
		}
	}
}

func TestSave_Multiple(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.yaml")
	r := newYAMLRepository(path)

	sessionA := &Session{Commands: []Command{{Name: "a", Payload: "QQ=="}}}
	sessionB := &Session{Commands: []Command{{Name: "b", Payload: "Qg=="}}}

	if err := r.Save(sessionA); err != nil {
		t.Fatal(err)
	}
	if err := r.Save(sessionB); err != nil {
		t.Fatal(err)
	}
	loaded, err := r.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Commands) != 1 || loaded.Commands[0].Name != "b" {
		t.Errorf("expected only session B, got %+v", loaded.Commands)
	}
}

func TestSave_EmptySession(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session.yaml")
	r := newYAMLRepository(path)
	if err := r.Save(&Session{}); err != nil {
		t.Fatal(err)
	}
	loaded, err := r.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Commands) != 0 {
		t.Errorf("expected empty commands, got %d", len(loaded.Commands))
	}
}
