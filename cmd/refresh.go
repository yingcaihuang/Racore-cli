package cmd

import (
	"fmt"
	"time"

	"racore-cli/internal/auth"
	"racore-cli/internal/credential"

	"github.com/spf13/cobra"
)

var refreshTokenCmd = &cobra.Command{
	Use:   "refresh-token",
	Short: "Force refresh the authentication token",
	Example: `  # Refresh token using stored credentials
  racore-cli refresh-token`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := credential.NewStore()
		if err != nil {
			return fmt.Errorf("cannot initialize credential store: %w", err)
		}

		creds, err := store.Load()
		if err != nil {
			return fmt.Errorf("cannot load credentials: %w", err)
		}

		if creds == nil {
			return fmt.Errorf("not logged in. Please run 'racore-cli login' first")
		}

		fmt.Println("Refreshing token...")

		// Re-authenticate to get a new token
		mgr := auth.NewManager(creds.AccessKey, creds.SecretKey)
		resp, err := mgr.Authenticate()
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}

		if resp.Code != 1 {
			return fmt.Errorf("authentication failed: %s", resp.Message)
		}

		// Update stored credentials with new token
		creds.Token = resp.Data.Token
		creds.Expire = resp.Data.Expire

		if err := store.Save(creds); err != nil {
			return fmt.Errorf("cannot save updated credentials: %w", err)
		}

		expireTime := time.Unix(resp.Data.Expire, 0).Format(time.RFC3339)
		remaining := resp.Data.Expire - time.Now().Unix()
		hours := remaining / 3600
		minutes := (remaining % 3600) / 60

		fmt.Printf("✓ Token refreshed successfully.\n")
		fmt.Printf("  Expires: %s (%dh %dm remaining)\n", expireTime, hours, minutes)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(refreshTokenCmd)
}
