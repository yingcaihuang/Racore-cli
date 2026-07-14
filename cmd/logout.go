package cmd

import (
	"fmt"
	"os"

	"racore-cli/internal/credential"

	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear stored credentials and revoke local access",
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
			fmt.Println("Already logged out")
			return nil
		}

		if err := store.Delete(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			os.Exit(1)
		}

		fmt.Println("Logged out successfully.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
