package core

import (
	"errors"
	"os"
	"path/filepath"
)

var ErrStateExists = errors.New("state already exists")

type State struct {
	path string
}

// NewState creates a new State from a specific root path.
func NewState(path string) (*State, error) {
	path = filepath.Join(filepath.Clean(path), ".casefile/")
	if err := os.Mkdir(path, 0755); err != nil {
		if os.IsExist(err) {
			return nil, ErrStateExists
		}
		return nil, err
	}
	return &State{path: path}, nil
}

func (s *State) Path() string {
	return s.path
}

func (s *State) AbsolutePath() string {
	absolutePath, err := filepath.Abs(s.path)
	if err != nil {
		return "-"
	}
	return absolutePath
}
