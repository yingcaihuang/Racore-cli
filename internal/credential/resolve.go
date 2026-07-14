package credential

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// InputSource holds credential flag values passed from the CLI layer.
type InputSource struct {
	FlagAccessKey string
	FlagSecretKey string
}

// Resolve returns accessKey, secretKey following priority: flags > env vars > interactive prompt.
func Resolve(src InputSource) (accessKey, secretKey string, err error) {
	// Priority 1: command-line flags
	if src.FlagAccessKey != "" && src.FlagSecretKey != "" {
		return src.FlagAccessKey, src.FlagSecretKey, nil
	}

	// Priority 2: environment variables
	envAccessKey := os.Getenv("RACORE_ACCESS_KEY")
	envSecretKey := os.Getenv("RACORE_SECRET_KEY")
	if envAccessKey != "" && envSecretKey != "" {
		return envAccessKey, envSecretKey, nil
	}

	// Priority 3: interactive prompt
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Access Key: ")
	accessKey, err = reader.ReadString('\n')
	if err != nil {
		return "", "", fmt.Errorf("failed to read access key: %w", err)
	}
	accessKey = strings.TrimSpace(accessKey)

	fmt.Print("Secret Key: ")
	secretBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", "", fmt.Errorf("failed to read secret key: %w", err)
	}
	fmt.Println() // newline after hidden input
	secretKey = string(secretBytes)

	if accessKey == "" || secretKey == "" {
		return "", "", fmt.Errorf("access key and secret key must not be empty")
	}

	return accessKey, secretKey, nil
}
