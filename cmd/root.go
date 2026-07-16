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
	Short:         "Racore Cloud CDN management CLI with MCP protocol support",
	Long: `Racore Cloud CDN management CLI with MCP protocol support.

Manage CDN domains, certificates, cache, statistics, work orders, and logs
through command line or AI agents via the MCP (Model Context Protocol).

If you encounter any bugs, please report them to: yingcai.huang@verycloud.cn`,
	Version:       "0.1.0",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		// Skip version check for non-interactive or self-referencing commands
		name := cmd.Name()
		if name == "serve" || name == "update" || name == "completion" || name == "help" {
			return
		}
		// Skip if not running in an interactive terminal
		if !isInteractiveTerminal() {
			return
		}
		checkForUpdate()
	}
}

// SetVersionInfo sets the version information from build-time ldflags.
func SetVersionInfo(version, commit, date string) {
	rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)
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
