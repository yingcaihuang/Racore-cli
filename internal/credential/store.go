package credential

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Credentials 表示持久化的凭证数据
type Credentials struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Token     string `json:"token,omitempty"`
	Expire    int64  `json:"expire,omitempty"`
}

// Store 管理凭证文件的读写
type Store struct {
	dir  string // ~/.racore/
	file string // ~/.racore/credentials
}

// NewStore 创建凭证存储实例
func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}
	dir := filepath.Join(home, ".racore")
	return &Store{
		dir:  dir,
		file: filepath.Join(dir, "credentials"),
	}, nil
}

// Save 持久化凭证到磁盘（0700 目录 + 0600 文件）
func (s *Store) Save(creds *Credentials) error {
	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return fmt.Errorf("cannot create credentials directory: %w", err)
	}

	data, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("cannot serialize credentials: %w", err)
	}

	if err := os.WriteFile(s.file, data, 0600); err != nil {
		return fmt.Errorf("cannot write credentials file: %w", err)
	}

	return nil
}

// Load 从磁盘加载凭证，文件不存在返回 (nil, nil)
func (s *Store) Load() (*Credentials, error) {
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

// Delete 删除凭证文件
func (s *Store) Delete() error {
	err := os.Remove(s.file)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("cannot delete credentials file: %w", err)
	}
	return nil
}

// CheckPermissions 检查文件权限，超过 0600 返回 warning
func (s *Store) CheckPermissions() (warning string, err error) {
	info, err := os.Stat(s.file)
	if err != nil {
		return "", err
	}
	mode := info.Mode().Perm()
	if mode&0077 != 0 {
		return fmt.Sprintf("WARNING: credentials file has permission %04o, expected 0600", mode), nil
	}
	return "", nil
}

// ClearSensitive 将 Credentials 中敏感字段对应的字节切片清零
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

// MaskAccessKey 掩码显示 access_key
// 长度 > 8 显示 first4 + "****" + last4，否则返回 "****"
func MaskAccessKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}
