package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	checkInterval = 24 * time.Hour
	repoAPI       = "https://api.github.com/repos/kno-ai/kno/releases/latest"
	stateFile     = "update-check"
)

type state struct {
	LastCheck     time.Time `json:"last_check"`
	LatestVersion string    `json:"latest_version"`
}

// Check compares the current version against the latest GitHub release.
// Returns a user-facing message if an update is available, or "" if up to date.
// Errors are swallowed — this should never break the CLI.
// Disabled by setting KNO_NO_UPDATE_CHECK=1.
func Check(currentVersion string) string {
	if currentVersion == "dev" {
		return ""
	}
	if os.Getenv("KNO_NO_UPDATE_CHECK") == "1" {
		return ""
	}

	dir := stateDir()
	if dir == "" {
		return ""
	}

	s := readState(dir)
	if time.Since(s.LastCheck) < checkInterval {
		return compareVersions(currentVersion, s.LatestVersion)
	}

	latest, err := fetchLatest()
	if err != nil {
		return ""
	}

	s.LastCheck = time.Now()
	s.LatestVersion = latest
	writeState(dir, s)

	return compareVersions(currentVersion, latest)
}

func compareVersions(current, latest string) string {
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")
	if latest == "" || current == latest {
		return ""
	}
	if !isNewer(latest, current) {
		return ""
	}
	return fmt.Sprintf("A new version of kno is available (%s → %s). Run: brew upgrade kno", current, latest)
}

// isNewer returns true if version a is strictly newer than version b.
// Compares major.minor.patch numerically. Returns false on parse errors.
func isNewer(a, b string) bool {
	ap := parseSemver(a)
	bp := parseSemver(b)
	if ap == nil || bp == nil {
		return false
	}
	for i := 0; i < 3; i++ {
		if ap[i] != bp[i] {
			return ap[i] > bp[i]
		}
	}
	return false
}

func parseSemver(v string) []int {
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return nil
	}
	nums := make([]int, 3)
	for i, p := range parts {
		// Strip pre-release suffix (e.g., "0-rc1")
		if idx := strings.IndexByte(p, '-'); idx >= 0 {
			p = p[:idx]
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		nums[i] = n
	}
	return nums
}

func fetchLatest() (string, error) {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(repoAPI)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	return strings.TrimPrefix(release.TagName, "v"), nil
}

func stateDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	dir := filepath.Join(home, ".kno")
	os.MkdirAll(dir, 0o755)
	return dir
}

func readState(dir string) state {
	data, err := os.ReadFile(filepath.Join(dir, stateFile))
	if err != nil {
		return state{}
	}
	var s state
	json.Unmarshal(data, &s)
	return s
}

func writeState(dir string, s state) {
	data, _ := json.Marshal(s)
	os.WriteFile(filepath.Join(dir, stateFile), data, 0o644)
}
