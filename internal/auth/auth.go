package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	AuthEndpoint          = "https://portal.racorecloud.com/API/OAuth/token"
	TokenRefreshThreshold = 300 // seconds
	RequestTimeout        = 30 * time.Second
)

// ComputeSignature 计算 HMAC-SHA512 签名
// message = dateStr + accessKey + secretKey
// key = secretKey
// 返回小写十六进制字符串
func ComputeSignature(accessKey, secretKey, dateStr string) string {
	message := dateStr + accessKey + secretKey
	mac := hmac.New(sha512.New, []byte(secretKey))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// FormatRFC1123 将时间格式化为 RFC1123 (HTTP date) 字符串
func FormatRFC1123(t time.Time) string {
	return t.UTC().Format(time.RFC1123)
}

// TokenResponse 表示 OAuth 接口返回的数据结构
type TokenResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Token  string `json:"token"`
		Expire int64  `json:"expire"`
	} `json:"data"`
}

// Manager 管理 Token 获取与缓存
type Manager struct {
	accessKey string
	secretKey string
	token     string
	expire    int64
	client    *http.Client
}

// NewManager 创建认证管理器
func NewManager(accessKey, secretKey string) *Manager {
	return &Manager{
		accessKey: accessKey,
		secretKey: secretKey,
		client: &http.Client{
			Timeout: RequestTimeout,
		},
	}
}

// SetCache 从已存储的凭证恢复缓存状态
func (m *Manager) SetCache(token string, expire int64) {
	m.token = token
	m.expire = expire
}

// GetValidToken 获取有效 Token，必要时自动刷新
func (m *Manager) GetValidToken() (string, error) {
	if m.IsTokenValid() {
		return m.token, nil
	}

	resp, err := m.Authenticate()
	if err != nil {
		return "", err
	}

	if resp.Code != 1 {
		return "", fmt.Errorf("authentication failed: %s", resp.Message)
	}

	return m.token, nil
}

// IsTokenValid 判断缓存 Token 是否有效（距过期 >= 300s）
func (m *Manager) IsTokenValid() bool {
	if m.token == "" {
		return false
	}
	return m.expire-time.Now().Unix() >= TokenRefreshThreshold
}

// Authenticate 执行认证请求获取新 Token
func (m *Manager) Authenticate() (*TokenResponse, error) {
	now := time.Now()
	dateStr := FormatRFC1123(now)
	signature := ComputeSignature(m.accessKey, m.secretKey, dateStr)

	// 构建请求 body
	body := map[string]string{
		"access_key": m.accessKey,
		"signature":  signature,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// 创建 POST 请求
	req, err := http.NewRequest(http.MethodPost, AuthEndpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("x-request-date", dateStr)
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("authentication request failed: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// 解析响应
	var tokenResp TokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 如果认证成功，更新缓存
	if tokenResp.Code == 1 {
		m.token = tokenResp.Data.Token
		m.expire = tokenResp.Data.Expire
	}

	return &tokenResp, nil
}

// ClearCache 清除 Token 缓存
func (m *Manager) ClearCache() {
	m.token = ""
	m.expire = 0
}
