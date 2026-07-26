package core

import (
	"casefile/internal/database"
	"errors"
	"os"
	"path/filepath"
)

var ErrStateExists = errors.New("state already exists")

type State struct {
	path string
	db   *database.Database
}

// NewState creates a new State from a specific root path.
func NewState(path string) (*State, error) {
	root := filepath.Join(filepath.Clean(path), ".casefile/")
	if err := os.Mkdir(root, 0755); err != nil {
		if os.IsExist(err) {
			return nil, ErrStateExists
		}
		return nil, err
	}

	// Set up the database.
	db, err := database.Connect(filepath.Join(root, "store.db"))
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("SELECT 1")
	if err != nil {
		return nil, err
	}

	return &State{path: path, db: db}, nil
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
