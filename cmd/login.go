package cmd

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"racore-cli/internal/auth"
	"racore-cli/internal/credential"

	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Racore Cloud and store credentials",
	RunE:  runLogin,
}

func init() {
	loginCmd.Flags().String("access-key", "", "Racore Cloud access key")
	loginCmd.Flags().String("secret-key", "", "Racore Cloud secret key")
	rootCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	ak, _ := cmd.Flags().GetString("access-key")
	sk, _ := cmd.Flags().GetString("secret-key")

	// Resolve credentials (flags → env → interactive prompt)
	accessKey, secretKey, err := credential.Resolve(credential.InputSource{
		FlagAccessKey: ak,
		FlagSecretKey: sk,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(3)
	}

	// Create a temporary Credentials struct for deferred cleanup
	creds := &credential.Credentials{
		AccessKey: accessKey,
		SecretKey: secretKey,
	}
	defer credential.ClearSensitive(creds)

	// Authenticate
	mgr := auth.NewManager(accessKey, secretKey)
	resp, err := mgr.Authenticate()
	if err != nil {
		// Check if it's a network error
		if isNetworkError(err) {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			os.Exit(2)
		}
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}

	// Check response code
	if resp.Code != 1 {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Message)
		return fmt.Errorf("%s", resp.Message)
	}

	// Save credentials with token and expire
	creds.Token = resp.Data.Token
	creds.Expire = resp.Data.Expire

	store, err := credential.NewStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return err
	}

	if err := store.Save(creds); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return err
	}

	// Output success message with formatted expire time
	expireTime := time.Unix(resp.Data.Expire, 0).Format(time.RFC3339)
	fmt.Printf("Login successful. Token expires at %s\n", expireTime)

	return nil
}

// isNetworkError checks if an error is a network-related error.
func isNetworkError(err error) bool {
	if err == nil {
		return false
	}
	// Check for net.Error (timeout, connection refused, DNS errors, etc.)
	if _, ok := err.(*net.OpError); ok {
		return true
	}
	// Check for wrapped network errors
	errMsg := err.Error()
	if strings.Contains(errMsg, "connection refused") ||
		strings.Contains(errMsg, "no such host") ||
		strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "dial tcp") ||
		strings.Contains(errMsg, "network is unreachable") {
		return true
	}
	return false
}
