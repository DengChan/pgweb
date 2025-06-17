package bookmarks

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/sosedoff/pgweb/pkg/shared"
)

// LastConnection represents the most recently used connection configuration
type LastConnection struct {
	Host     string          `toml:"host"`
	Port     int             `toml:"port"`
	User     string          `toml:"user"`
	Database string          `toml:"database"`
	SSLMode  string          `toml:"ssl_mode"`
	SSH      *shared.SSHInfo `toml:"ssh,omitempty"`
	LastUsed time.Time       `toml:"last_used"`
}

// LastConnectionManager manages the last used connection configuration
type LastConnectionManager struct {
	dir string
}

// NewLastConnectionManager creates a new LastConnectionManager
func NewLastConnectionManager(dir string) *LastConnectionManager {
	return &LastConnectionManager{
		dir: dir,
	}
}

// Save saves the last connection configuration
func (m *LastConnectionManager) Save(conn *LastConnection) error {
	if m.dir == "" {
		return fmt.Errorf("bookmarks directory not set")
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(m.dir, 0755); err != nil {
		return fmt.Errorf("failed to create bookmarks directory: %w", err)
	}

	conn.LastUsed = time.Now()

	// Save to a special file
	filePath := filepath.Join(m.dir, "last_connection.toml")

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create last connection file: %w", err)
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	if err := encoder.Encode(conn); err != nil {
		return fmt.Errorf("failed to encode last connection: %w", err)
	}

	return nil
}

// Load loads the last connection configuration
func (m *LastConnectionManager) Load() (*LastConnection, error) {
	if m.dir == "" {
		return nil, fmt.Errorf("bookmarks directory not set")
	}

	filePath := filepath.Join(m.dir, "last_connection.toml")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, nil // No last connection saved
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read last connection file: %w", err)
	}

	var conn LastConnection
	if _, err := toml.Decode(string(data), &conn); err != nil {
		return nil, fmt.Errorf("failed to decode last connection: %w", err)
	}

	// Set default port if not provided
	if conn.Port == 0 {
		conn.Port = 5432
	}

	// Set default SSL mode if not provided
	if conn.SSLMode == "" {
		conn.SSLMode = "disable"
	}

	return &conn, nil
}

// Clear removes the last connection configuration
func (m *LastConnectionManager) Clear() error {
	if m.dir == "" {
		return fmt.Errorf("bookmarks directory not set")
	}

	filePath := filepath.Join(m.dir, "last_connection.toml")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to clear
	}

	return os.Remove(filePath)
}
