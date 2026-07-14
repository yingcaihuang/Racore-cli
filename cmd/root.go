package cmd

import (
	"errors"
	"fmt"
	"os"

	"racore-cli/internal/api"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "racore-cli",
	Short:         "Racore Cloud CDN management CLI with ACP protocol support",
	Version:       "0.1.0",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command and handles errors.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())

		// Determine exit code based on error type
		var authErr *AuthError
		var apiAuthErr *api.AuthError
		var netErr *NetworkError
		var inputErr *InputError

		switch {
		case errors.As(err, &authErr):
			os.Exit(1)
		case errors.As(err, &apiAuthErr):
			os.Exit(1)
		case errors.As(err, &netErr):
			os.Exit(2)
		case errors.As(err, &inputErr):
			os.Exit(3)
		default:
			os.Exit(1)
		}
	}
}
