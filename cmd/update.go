package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update racore-cli to the latest version",
	Example: `  # Check and update to latest version
  racore-cli update

  # Force self-update (skip package manager detection)
  racore-cli update --force`,
	RunE: runUpdate,
}

func init() {
	updateCmd.Flags().Bool("force", false, "Force self-update, skip package manager detection")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")
	currentVersion := strings.TrimPrefix(rootCmd.Version, "")

	// Extract just the semver part (e.g., "0.2.5" from "0.2.5 (commit: xxx, built: xxx)")
	if idx := strings.Index(currentVersion, " "); idx > 0 {
		currentVersion = currentVersion[:idx]
	}

	fmt.Printf("Current version: %s\n", currentVersion)
	fmt.Println("Checking for updates...")

	// Try package manager first (unless --force)
	if !force {
		if updated := tryPackageManagerUpdate(); updated {
			return nil
		}
	}

	// Fallback: self-update from GitHub Releases
	return selfUpdate(currentVersion)
}

// tryPackageManagerUpdate detects the installation method and uses the appropriate package manager
func tryPackageManagerUpdate() bool {
	switch runtime.GOOS {
	case "darwin":
		if isInstalledViaBrew() {
			fmt.Println("Detected Homebrew installation. Running brew upgrade...")
			cmd := exec.Command("brew", "upgrade", "racore-cli")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Homebrew upgrade failed: %v\nFalling back to self-update...\n", err)
				return false
			}
			fmt.Println("✓ Updated via Homebrew successfully.")
			return true
		}
	case "windows":
		if isInstalledViaScoop() {
			fmt.Println("Detected Scoop installation. Running scoop update...")
			cmd := exec.Command("scoop", "update", "racore-cli")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Scoop update failed: %v\nFalling back to self-update...\n", err)
				return false
			}
			fmt.Println("✓ Updated via Scoop successfully.")
			return true
		}
	case "linux":
		if isInstalledViaDeb() {
			fmt.Println("Detected dpkg installation. Downloading latest .deb...")
			return updateViaDeb()
		}
		if isInstalledViaRpm() {
			fmt.Println("Detected rpm installation. Downloading latest .rpm...")
			return updateViaRpm()
		}
	}
	return false
}

// selfUpdate downloads and replaces the binary from GitHub Releases
func selfUpdate(currentVersion string) error {
	source, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
	if err != nil {
		return fmt.Errorf("failed to create update source: %w", err)
	}

	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source: source,
	})
	if err != nil {
		return fmt.Errorf("failed to create updater: %w", err)
	}

	latest, found, err := updater.DetectLatest(context.Background(), selfupdate.ParseSlug("yingcaihuang/Racore-cli"))
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}
	if !found {
		return fmt.Errorf("no releases found")
	}

	if latest.LessOrEqual(currentVersion) {
		fmt.Printf("✓ Already up to date (v%s).\n", currentVersion)
		return nil
	}

	fmt.Printf("New version available: v%s → v%s\n", currentVersion, latest.Version())
	fmt.Println("Downloading and installing...")

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot determine executable path: %w", err)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return fmt.Errorf("cannot resolve executable path: %w", err)
	}

	if err := updater.UpdateTo(context.Background(), latest, exe); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	fmt.Printf("✓ Successfully updated to v%s\n", latest.Version())
	return nil
}

// --- Detection helpers ---

func isInstalledViaBrew() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return false
	}
	// Homebrew installs to /opt/homebrew/Cellar/ or /usr/local/Cellar/
	return strings.Contains(resolved, "/Cellar/")
}

func isInstalledViaScoop() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	// Scoop installs to ~/scoop/apps/
	return strings.Contains(strings.ToLower(exe), "scoop")
}

func isInstalledViaDeb() bool {
	_, err := exec.LookPath("dpkg")
	if err != nil {
		return false
	}
	cmd := exec.Command("dpkg", "-s", "racore-cli")
	return cmd.Run() == nil
}

func isInstalledViaRpm() bool {
	_, err := exec.LookPath("rpm")
	if err != nil {
		return false
	}
	cmd := exec.Command("rpm", "-q", "racore-cli")
	return cmd.Run() == nil
}

func updateViaDeb() bool {
	// Get latest version
	latestVersion := getLatestVersionTag()
	if latestVersion == "" {
		fmt.Fprintln(os.Stderr, "Failed to determine latest version.")
		return false
	}

	arch := runtime.GOARCH
	url := fmt.Sprintf("https://github.com/yingcaihuang/Racore-cli/releases/download/v%s/racore-cli_%s_linux_%s.deb", latestVersion, latestVersion, arch)

	tmpFile := fmt.Sprintf("/tmp/racore-cli_%s_linux_%s.deb", latestVersion, arch)
	cmd := exec.Command("curl", "-fsSL", "-o", tmpFile, url)
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		return false
	}

	cmd = exec.Command("sudo", "dpkg", "-i", tmpFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "dpkg install failed: %v\n", err)
		return false
	}

	os.Remove(tmpFile)
	fmt.Printf("✓ Updated via dpkg to v%s\n", latestVersion)
	return true
}

func updateViaRpm() bool {
	latestVersion := getLatestVersionTag()
	if latestVersion == "" {
		fmt.Fprintln(os.Stderr, "Failed to determine latest version.")
		return false
	}

	arch := runtime.GOARCH
	url := fmt.Sprintf("https://github.com/yingcaihuang/Racore-cli/releases/download/v%s/racore-cli_%s_linux_%s.rpm", latestVersion, latestVersion, arch)

	tmpFile := fmt.Sprintf("/tmp/racore-cli_%s_linux_%s.rpm", latestVersion, arch)
	cmd := exec.Command("curl", "-fsSL", "-o", tmpFile, url)
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		return false
	}

	cmd = exec.Command("sudo", "rpm", "-U", tmpFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "rpm upgrade failed: %v\n", err)
		return false
	}

	os.Remove(tmpFile)
	fmt.Printf("✓ Updated via rpm to v%s\n", latestVersion)
	return true
}

func getLatestVersionTag() string {
	cmd := exec.Command("curl", "-fsSL", "https://api.github.com/repos/yingcaihuang/Racore-cli/releases/latest")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	// Simple parse: find "tag_name": "v0.x.x"
	s := string(output)
	idx := strings.Index(s, `"tag_name"`)
	if idx < 0 {
		return ""
	}
	s = s[idx:]
	start := strings.Index(s, `"v`)
	if start < 0 {
		return ""
	}
	s = s[start+2:]
	end := strings.Index(s, `"`)
	if end < 0 {
		return ""
	}
	return s[:end]
}
