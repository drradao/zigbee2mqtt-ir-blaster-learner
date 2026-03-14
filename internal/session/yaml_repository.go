package session

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type yamlRepository struct {
	path string
}

func newYAMLRepository(path string) *yamlRepository {
	return &yamlRepository{path: path}
}

// Load reads the session YAML. A missing file is not an error.
func (r *yamlRepository) Load() (*Session, error) {
	s := &Session{}
	data, err := os.ReadFile(r.path)
	if errors.Is(err, fs.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, s); err != nil {
		return nil, err
	}
	return s, nil
}

// Save writes the session atomically via a temp file rename.
func (r *yamlRepository) Save(s *Session) error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "zigbee2mqtt-ir-blaster-learner-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, r.path); err != nil {
		os.Remove(tmpName)
		return err
	}
	return nil
}
