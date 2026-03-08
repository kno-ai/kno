package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// mcpClient describes a known MCP client that kno can register with.
type mcpClient struct {
	Name       string // user-facing name for --register flag
	ConfigPath string // absolute path to config file
}

// knownMCPClients returns all known MCP client configurations on this system.
func knownMCPClients() []mcpClient {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	var clients []mcpClient

	if runtime.GOOS == "darwin" {
		clients = append(clients, mcpClient{
			Name:       "claude-desktop",
			ConfigPath: filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json"),
		})
	}

	// Claude Code — cross-platform
	clients = append(clients, mcpClient{
		Name:       "claude-code",
		ConfigPath: filepath.Join(home, ".claude.json"),
	})

	// Filter to clients whose parent directory exists.
	var found []mcpClient
	for _, c := range clients {
		if info, err := os.Stat(filepath.Dir(c.ConfigPath)); err == nil && info.IsDir() {
			found = append(found, c)
		}
	}
	return found
}

// registerMCPClients registers kno with the given MCP clients.
// Returns the list of clients that were successfully registered and any errors.
func registerMCPClients(clients []mcpClient, vaultPath, serverName string) ([]mcpClient, []error) {
	var registered []mcpClient
	var errs []error
	for _, c := range clients {
		if err := registerMCPAt(c.ConfigPath, vaultPath, serverName); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", c.Name, err))
		} else {
			registered = append(registered, c)
		}
	}
	return registered, errs
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
		"type":    "stdio",
		"command": knoBin,
		"args":    []string{"--vault", vaultPath, "mcp"},
		"env":     map[string]string{},
	}
	clientConfig["mcpServers"] = servers

	out, err := json.MarshalIndent(clientConfig, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return err
	}

	out = append(out, '\n')
	return os.WriteFile(configPath, out, 0o644)
}
