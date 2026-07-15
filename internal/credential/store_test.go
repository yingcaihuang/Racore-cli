package credential

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewStore(t *testing.T) {
	store, err := NewStore()
	if err != nil {
		t.Fatalf("NewStore() returned error: %v", err)
	}
	if store == nil {
		t.Fatal("NewStore() returned nil store")
	}
	home, _ := os.UserHomeDir()
	expectedDir := filepath.Join(home, ".racore")
	if store.dir != expectedDir {
		t.Errorf("store.dir = %q, want %q", store.dir, expectedDir)
	}
	expectedFile := filepath.Join(home, ".racore", "credentials")
	if store.file != expectedFile {
		t.Errorf("store.file = %q, want %q", store.file, expectedFile)
	}
}

// newFileOnlyStore creates a Store that uses file backend only (for testing without keyring dependency)
func newFileOnlyStore(dir string) *Store {
	return &Store{
		useKeyring: false,
		dir:        dir,
		file:       filepath.Join(dir, "credentials"),
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	store := newFileOnlyStore(tmpDir)

	creds := &Credentials{
		AccessKey: "test-access-key-12345",
		SecretKey: "test-secret-key-67890",
		Token:     "test-token-abc",
		Expire:    1719000000,
	}

	// Save
	if err := store.Save(creds); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	// Verify file permissions
	info, err := os.Stat(store.file)
	if err != nil {
		t.Fatalf("cannot stat credentials file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("file permission = %04o, want 0600", perm)
	}

	// Load
	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if loaded == nil {
		t.Fatal("Load() returned nil")
	}
	if loaded.AccessKey != creds.AccessKey {
		t.Errorf("AccessKey = %q, want %q", loaded.AccessKey, creds.AccessKey)
	}
	if loaded.SecretKey != creds.SecretKey {
		t.Errorf("SecretKey = %q, want %q", loaded.SecretKey, creds.SecretKey)
	}
	if loaded.Token != creds.Token {
		t.Errorf("Token = %q, want %q", loaded.Token, creds.Token)
	}
	if loaded.Expire != creds.Expire {
		t.Errorf("Expire = %d, want %d", loaded.Expire, creds.Expire)
	}
}

func TestLoadNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	store := newFileOnlyStore(tmpDir)
	store.file = filepath.Join(tmpDir, "nonexistent")

	creds, err := store.Load()
	if err != nil {
		t.Fatalf("Load() returned error for nonexistent file: %v", err)
	}
	if creds != nil {
		t.Errorf("Load() returned non-nil for nonexistent file: %+v", creds)
	}
}

func TestDelete(t *testing.T) {
	tmpDir := t.TempDir()
	store := newFileOnlyStore(tmpDir)

	// Save first
	creds := &Credentials{AccessKey: "key", SecretKey: "secret"}
	if err := store.Save(creds); err != nil {
		t.Fatalf("Save() returned error: %v", err)
	}

	// Delete
	if err := store.Delete(); err != nil {
		t.Fatalf("Delete() returned error: %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(store.file); !os.IsNotExist(err) {
		t.Error("credentials file still exists after Delete()")
	}
}

func TestDeleteNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	store := newFileOnlyStore(tmpDir)
	store.file = filepath.Join(tmpDir, "nonexistent")

	// Should not return error for nonexistent file
	if err := store.Delete(); err != nil {
		t.Fatalf("Delete() returned error for nonexistent file: %v", err)
	}
}

func TestCheckPermissions_Secure(t *testing.T) {
	tmpDir := t.TempDir()
	store := newFileOnlyStore(tmpDir)

	// Write with 0600 permissions
	if err := os.WriteFile(store.file, []byte("{}"), 0600); err != nil {
		t.Fatalf("cannot create test file: %v", err)
	}

	warning, err := store.CheckPermissions()
	if err != nil {
		t.Fatalf("CheckPermissions() returned error: %v", err)
	}
	if warning != "" {
		t.Errorf("CheckPermissions() returned warning for 0600 file: %q", warning)
	}
}

func TestCheckPermissions_Insecure(t *testing.T) {
	tmpDir := t.TempDir()
	store := newFileOnlyStore(tmpDir)

	// Write with overly permissive permissions
	if err := os.WriteFile(store.file, []byte("{}"), 0644); err != nil {
		t.Fatalf("cannot create test file: %v", err)
	}

	warning, err := store.CheckPermissions()
	if err != nil {
		t.Fatalf("CheckPermissions() returned error: %v", err)
	}
	if warning == "" {
		t.Error("CheckPermissions() returned no warning for 0644 file")
	}
}

func TestCheckPermissions_Keyring(t *testing.T) {
	tmpDir := t.TempDir()
	store := &Store{
		useKeyring: true,
		dir:        tmpDir,
		file:       filepath.Join(tmpDir, "credentials"),
	}

	// When using keyring, CheckPermissions should always return no warning
	warning, err := store.CheckPermissions()
	if err != nil {
		t.Fatalf("CheckPermissions() returned error: %v", err)
	}
	if warning != "" {
		t.Errorf("CheckPermissions() returned warning when using keyring: %q", warning)
	}
}

func TestStorageType(t *testing.T) {
	store := &Store{useKeyring: true}
	if got := store.StorageType(); got != "system keyring" {
		t.Errorf("StorageType() = %q, want %q", got, "system keyring")
	}

	store = &Store{useKeyring: false}
	if got := store.StorageType(); got != "file (~/.racore/credentials)" {
		t.Errorf("StorageType() = %q, want %q", got, "file (~/.racore/credentials)")
	}
}

func TestClearSensitive(t *testing.T) {
	creds := &Credentials{
		AccessKey: "my-access-key",
		SecretKey: "my-secret-key",
		Token:     "my-token",
		Expire:    12345,
	}

	ClearSensitive(creds)

	if creds.SecretKey != "" {
		t.Errorf("SecretKey not cleared: %q", creds.SecretKey)
	}
	if creds.Token != "" {
		t.Errorf("Token not cleared: %q", creds.Token)
	}
	// AccessKey should remain unchanged
	if creds.AccessKey != "my-access-key" {
		t.Errorf("AccessKey was modified: %q", creds.AccessKey)
	}
}

func TestClearSensitive_Nil(t *testing.T) {
	// Should not panic on nil
	ClearSensitive(nil)
}

func TestMaskAccessKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", "****"},
		{"abcd", "****"},
		{"abcdefgh", "****"},           // exactly 8
		{"abcdefghi", "abcd****fghi"},  // 9 characters
		{"AKID12345678ABCD", "AKID****ABCD"},
		{"123456789", "1234****6789"},
	}

	for _, tt := range tests {
		got := MaskAccessKey(tt.input)
		if got != tt.want {
			t.Errorf("MaskAccessKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
