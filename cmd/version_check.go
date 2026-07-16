package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	checkInterval    = 24 * time.Hour
	githubReleaseURL = "https://api.github.com/repos/yingcaihuang/Racore-cli/releases/latest"
)

type versionCache struct {
	LatestVersion string `json:"latest_version"`
	CheckedAt     int64  `json:"checked_at"`
}

// checkForUpdate checks if a newer version is available and prompts the user.
// It caches the result to avoid checking more than once per day.
func checkForUpdate() {
	currentVersion := getCurrentVersion()
	if currentVersion == "" || currentVersion == "dev" {
		return
	}

	latestVersion := getLatestVersionCached()
	if latestVersion == "" {
		return
	}

	if !isNewer(latestVersion, currentVersion) {
		return
	}

	// Prompt user
	fmt.Fprintf(os.Stderr, "\nracore-cli %s available (current %s). Upgrade now? [y/N]: ", latestVersion, currentVersion)

	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer == "y" || answer == "yes" {
		fmt.Fprintf(os.Stderr, "Updating racore-cli %s → %s ...\n", currentVersion, latestVersion)
		cmd := exec.Command(os.Args[0], "update")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Run()
	}
}

// getCurrentVersion extracts the semver from rootCmd.Version
func getCurrentVersion() string {
	v := rootCmd.Version
	if idx := strings.Index(v, " "); idx > 0 {
		v = v[:idx]
	}
	return v
}

// getLatestVersionCached returns the latest version, using cache if fresh enough
func getLatestVersionCached() string {
	cache := loadVersionCache()

	// If cache is fresh (less than 24h old), use it
	if cache != nil && time.Since(time.Unix(cache.CheckedAt, 0)) < checkInterval {
		return cache.LatestVersion
	}

	// Fetch from GitHub (with short timeout)
	latestVersion := fetchLatestVersion()
	if latestVersion != "" {
		saveVersionCache(&versionCache{
			LatestVersion: latestVersion,
			CheckedAt:     time.Now().Unix(),
		})
	}

	return latestVersion
}

// fetchLatestVersion fetches the latest release version from GitHub
func fetchLatestVersion() string {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(githubReleaseURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ""
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return ""
	}

	return strings.TrimPrefix(release.TagName, "v")
}

// isNewer returns true if latest is newer than current (semver comparison)
func isNewer(latest, current string) bool {
	latestParts := strings.Split(latest, ".")
	currentParts := strings.Split(current, ".")

	for i := 0; i < len(latestParts) && i < len(currentParts); i++ {
		if latestParts[i] > currentParts[i] {
			return true
		}
		if latestParts[i] < currentParts[i] {
			return false
		}
	}
	return len(latestParts) > len(currentParts)
}

// isInteractiveTerminal checks if stdin is a terminal (not piped)
func isInteractiveTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// --- Cache file helpers (stored in ~/.racore/version_check.json) ---

func versionCacheFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".racore", "version_check.json")
}

func loadVersionCache() *versionCache {
	file := versionCacheFile()
	if file == "" {
		return nil
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return nil
	}
	var cache versionCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil
	}
	return &cache
}

func saveVersionCache(cache *versionCache) {
	file := versionCacheFile()
	if file == "" {
		return
	}
	dir := filepath.Dir(file)
	os.MkdirAll(dir, 0700)
	data, err := json.Marshal(cache)
	if err != nil {
		return
	}
	os.WriteFile(file, data, 0600)
}
