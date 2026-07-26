package core

import (
	"casefile/internal/database"
	"errors"
	"os"
	"path/filepath"
)

var ErrStateExists = errors.New("state already exists")

type State struct {
	path   string
	config *Config
	db     *database.Database
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

	// Create the default configuration file at the working directory.
	config, err := NewConfig(path, "Default")

	// Set up the database.
	db, err := database.Connect(filepath.Join(root, "store.db"))
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("SELECT 1")
	if err != nil {
		return nil, err
	}

	return &State{
		path:   root,
		config: config,
		db:     db,
	}, nil
}

func (s *State) Config() *Config {
	return s.config
}

func (s *State) DB() *database.Database {
	return s.db
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
