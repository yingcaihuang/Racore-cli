package cmd

import (
	"fmt"
	"os"
	"time"

	"racore-cli/internal/credential"

	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Display current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := credential.NewStore()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			os.Exit(1)
		}

		creds, err := store.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			os.Exit(1)
		}

		if creds == nil {
			fmt.Fprintf(os.Stderr, "Error: Not logged in\n")
			os.Exit(1)
		}

		// Check file permissions and warn if too permissive
		if warning, err := store.CheckPermissions(); err == nil && warning != "" {
			fmt.Fprintf(os.Stderr, "%s\n", warning)
		}

		// Display masked access key
		fmt.Printf("Access Key: %s\n", credential.MaskAccessKey(creds.AccessKey))

		// Calculate token status
		remaining := creds.Expire - time.Now().Unix()
		if remaining >= 300 {
			hours := remaining / 3600
			minutes := (remaining % 3600) / 60
			fmt.Printf("Token Status: Valid (%dh %dm remaining)\n", hours, minutes)
		} else {
			fmt.Printf("Token Status: Expired/Expiring\n")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}
