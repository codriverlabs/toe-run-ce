package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Manager struct {
	basePath string
}

func NewManager(basePath string) (*Manager, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &Manager{
		basePath: basePath,
	}, nil
}

func (m *Manager) SaveProfile(r io.Reader, filename string) error {
	path := filepath.Join(m.basePath, filename)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create profile file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			// Log error but don't fail the operation
			_ = err
		}
	}()

	if _, err := io.Copy(f, r); err != nil {
		return fmt.Errorf("failed to write profile data: %w", err)
	}

	return nil
}
