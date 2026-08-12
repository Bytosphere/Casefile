package core

import (
	"casefile/internal/core/tool"
	"casefile/internal/database"
	"errors"
	"os"
	"path/filepath"
)

var ErrStateExists = errors.New("state already exists")

type State struct {
	path    string
	profile *Profile
	db      *database.Database
	tools   *tool.Registry
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

	// Load the default profile. If it doesn't exist, create it.
	profile, err := LoadProfile(path, "default")
	if err != nil {
		// If no profile exists, create the default configuration.
		if errors.Is(err, ErrProfileNotFound) {
			_, err = NewConfig(path, "default")
			if err != nil && !errors.Is(err, ErrConfigExists) {
				return nil, err
			}
			profile, err = LoadProfile(path, "default")
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// Set up the database.
	db, err := database.Connect(filepath.Join(root, "store.db"))
	if err != nil {
		return nil, err
	}

	// TODO: Test query. Remove later.
	_, err = db.Exec("SELECT 1")
	if err != nil {
		return nil, err
	}

	return &State{
		path:    root,
		profile: profile,
		db:      db,
		tools:   tool.NewRegistry(),
	}, nil
}

// LoadState loads the current state of the project.
func LoadState() (*State, error) {
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Append the State's path.
	path := filepath.Join(filepath.Clean(root), ".casefile/")

	profile, err := LoadProfile(root, "default")
	if err != nil {
		return nil, err
	}

	db, err := database.Connect(filepath.Join(path, "store.db"))
	if err != nil {
		return nil, err
	}

	return &State{
		path:    path,
		profile: profile,
		db:      db,
		tools:   tool.NewRegistry(),
	}, nil
}

func (s *State) Config() *Config {
	return &s.profile.config
}

func (s *State) Profile() *Profile {
	return s.profile
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
