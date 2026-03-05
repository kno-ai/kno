package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

// knownMCPConfigs returns paths to known MCP client config files
// that exist on this system.
func knownMCPConfigs() []string {
	if runtime.GOOS != "darwin" {
		return nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	candidates := []string{
		filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json"),
		// Future: Cursor, Windsurf, etc.
	}

	var found []string
	for _, c := range candidates {
		// Check if the parent directory exists (i.e. the app is installed).
		if info, err := os.Stat(filepath.Dir(c)); err == nil && info.IsDir() {
			found = append(found, c)
		}
	}
	return found
}

// registerMCP registers kno as an MCP server with the given config path,
// or auto-detects installed clients if path is empty.
// Returns the list of config files that were successfully updated.
func registerMCP(explicitPath string) []string {
	var targets []string
	if explicitPath != "" {
		targets = []string{explicitPath}
	} else {
		targets = knownMCPConfigs()
	}

	var registered []string
	for _, path := range targets {
		if err := registerMCPAt(path); err == nil {
			registered = append(registered, path)
		}
	}
	return registered
}

func registerMCPAt(configPath string) error {
	// Load existing config or start fresh.
	var clientConfig map[string]any
	data, err := os.ReadFile(configPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		clientConfig = make(map[string]any)
	} else {
		if err := json.Unmarshal(data, &clientConfig); err != nil {
			return err
		}
	}

	// Resolve the kno binary path.
	knoBin, err := os.Executable()
	if err != nil {
		knoBin = "kno"
	}

	// Ensure mcpServers key exists.
	servers, ok := clientConfig["mcpServers"].(map[string]any)
	if !ok {
		servers = make(map[string]any)
	}

	servers["kno"] = map[string]any{
		"command": knoBin,
		"args":    []string{"mcp"},
	}
	clientConfig["mcpServers"] = servers

	out, err := json.MarshalIndent(clientConfig, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return err
	}

	return os.WriteFile(configPath, out, 0o644)
}
