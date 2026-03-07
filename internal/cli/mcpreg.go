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
	}

	var found []string
	for _, c := range candidates {
		if info, err := os.Stat(filepath.Dir(c)); err == nil && info.IsDir() {
			found = append(found, c)
		}
	}
	return found
}

// registerMCP registers kno as an MCP server.
// Returns the list of config files that were successfully updated.
func registerMCP(explicitPath, vaultPath, serverName string) []string {
	var targets []string
	if explicitPath != "" {
		targets = []string{explicitPath}
	} else {
		targets = knownMCPConfigs()
	}

	var registered []string
	for _, path := range targets {
		if err := registerMCPAt(path, vaultPath, serverName); err == nil {
			registered = append(registered, path)
		}
	}
	return registered
}

func registerMCPAt(configPath, vaultPath, serverName string) error {
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

	knoBin, err := os.Executable()
	if err != nil {
		knoBin = "kno"
	}

	servers, ok := clientConfig["mcpServers"].(map[string]any)
	if !ok {
		servers = make(map[string]any)
	}

	servers[serverName] = map[string]any{
		"command": knoBin,
		"args":    []string{"--vault", vaultPath, "mcp"},
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
