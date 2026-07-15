package credential

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "racore-cli"
	accountName = "default"
)

// Credentials represents persisted credential data
type Credentials struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Token     string `json:"token,omitempty"`
	Expire    int64  `json:"expire,omitempty"`
}

// Store manages credential storage with OS keyring (preferred) or file fallback
type Store struct {
	useKeyring bool
	dir        string // ~/.racore/
	file       string // ~/.racore/credentials
}

// NewStore creates a credential store instance.
// It attempts to use the OS keyring and falls back to file storage if unavailable.
func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}
	dir := filepath.Join(home, ".racore")

	// Test if keyring is available by trying a no-op
	useKeyring := isKeyringAvailable()

	return &Store{
		useKeyring: useKeyring,
		dir:        dir,
		file:       filepath.Join(dir, "credentials"),
	}, nil
}

// isKeyringAvailable checks if the system keyring is accessible
func isKeyringAvailable() bool {
	// Try to get a non-existent key - if we get ErrNotFound, keyring works
	// If we get a different error, keyring is not available
	_, err := keyring.Get(serviceName, "__probe__")
	if err == keyring.ErrNotFound {
		return true
	}
	if err == nil {
		return true
	}
	// Any other error means keyring is not available (headless server, etc.)
	return false
}

// Save persists credentials to the OS keyring or file fallback
func (s *Store) Save(creds *Credentials) error {
	data, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("cannot serialize credentials: %w", err)
	}

	if s.useKeyring {
		if err := keyring.Set(serviceName, accountName, string(data)); err != nil {
			// Keyring failed, fall back to file
			return s.saveToFile(data)
		}
		// Also remove old file if it exists (migrating to keyring)
		os.Remove(s.file)
		return nil
	}

	return s.saveToFile(data)
}

// Load retrieves credentials from OS keyring or file fallback
func (s *Store) Load() (*Credentials, error) {
	if s.useKeyring {
		secret, err := keyring.Get(serviceName, accountName)
		if err == nil {
			var creds Credentials
			if err := json.Unmarshal([]byte(secret), &creds); err != nil {
				return nil, fmt.Errorf("cannot parse credentials from keyring: %w", err)
			}
			return &creds, nil
		}
		if err != keyring.ErrNotFound {
			// Keyring error, try file fallback
			return s.loadFromFile()
		}
		// Not found in keyring, try file (maybe not yet migrated)
		return s.loadFromFile()
	}

	return s.loadFromFile()
}

// Delete removes credentials from OS keyring and file
func (s *Store) Delete() error {
	var lastErr error

	if s.useKeyring {
		err := keyring.Delete(serviceName, accountName)
		if err != nil && err != keyring.ErrNotFound {
			lastErr = err
		}
	}

	// Also remove file if it exists
	err := os.Remove(s.file)
	if err != nil && !os.IsNotExist(err) {
		lastErr = err
	}

	if lastErr != nil {
		return fmt.Errorf("cannot delete credentials: %w", lastErr)
	}
	return nil
}

// CheckPermissions checks if credentials are stored securely
func (s *Store) CheckPermissions() (warning string, err error) {
	if s.useKeyring {
		return "", nil // Keyring is always secure
	}

	info, err := os.Stat(s.file)
	if err != nil {
		return "", err
	}
	mode := info.Mode().Perm()
	if mode&0077 != 0 {
		return fmt.Sprintf("WARNING: credentials file has permission %04o, expected 0600. Consider running 'chmod 600 %s'", mode, s.file), nil
	}
	return "", nil
}

// StorageType returns a description of the current storage backend
func (s *Store) StorageType() string {
	if s.useKeyring {
		return "system keyring"
	}
	return "file (~/.racore/credentials)"
}

// --- File fallback (for headless servers without keyring) ---

func (s *Store) saveToFile(data []byte) error {
	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return fmt.Errorf("cannot create credentials directory: %w", err)
	}
	if err := os.WriteFile(s.file, data, 0600); err != nil {
		return fmt.Errorf("cannot write credentials file: %w", err)
	}
	return nil
}

func (s *Store) loadFromFile() (*Credentials, error) {
	data, err := os.ReadFile(s.file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("cannot read credentials file: %w", err)
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("cannot parse credentials file: %w", err)
	}
	return &creds, nil
}

// ClearSensitive zeroes out sensitive fields in memory
func ClearSensitive(creds *Credentials) {
	if creds == nil {
		return
	}
	clear := func(s *string) {
		b := []byte(*s)
		for i := range b {
			b[i] = 0
		}
		*s = ""
	}
	clear(&creds.SecretKey)
	clear(&creds.Token)
}

// MaskAccessKey masks the access key for display
// Shows first4 + "****" + last4 if long enough, otherwise "****"
func MaskAccessKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}
