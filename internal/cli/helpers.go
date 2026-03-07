package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/kno-ai/kno/internal/app"
	"github.com/kno-ai/kno/internal/model"
	"github.com/spf13/cobra"
)

func defaultVaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return home + "/kno"
}

// resolveVault determines vault path from flag or default.
func resolveVault(cmd *cobra.Command) string {
	if v, _ := cmd.Flags().GetString("vault"); v != "" {
		return v
	}
	return defaultVaultPath()
}

// loadApp creates an App from the resolved vault path.
func loadApp(cmd *cobra.Command) (*app.App, error) {
	vaultPath := resolveVault(cmd)
	if vaultPath == "" {
		return nil, fmt.Errorf("could not determine vault path; use --vault or run 'kno setup'")
	}

	// Check vault exists
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("vault not found at %s; run 'kno setup' first", vaultPath)
	}

	return app.FromVaultPath(vaultPath)
}

// printJSON writes a JSON-encoded value to stdout.
func printJSON(w io.Writer, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(w, string(data))
	return nil
}

// metaMapToJSON converts a MetaMap to a map[string]any for JSON embedding.
// Single-value keys become scalars; multi-value keys become arrays.
func metaMapToJSON(m model.MetaMap) map[string]any {
	out := make(map[string]any)
	for k, vs := range m {
		if len(vs) == 1 {
			out[k] = vs[0]
		} else {
			out[k] = vs
		}
	}
	return out
}

// noteMetaJSON converts a MetaMap for note JSON output,
// ensuring curated_at and curated_into are always present (null if absent).
func noteMetaJSON(m model.MetaMap) map[string]any {
	out := metaMapToJSON(m)
	if _, ok := out["curated_at"]; !ok {
		out["curated_at"] = nil
	}
	if _, ok := out["curated_into"]; !ok {
		out["curated_into"] = nil
	}
	return out
}

// decorativeHeader builds a ━━━ padded header line to the target rune width.
func decorativeHeader(prefix string, targetWidth int) string {
	runes := []rune(prefix)
	for len(runes) < targetWidth {
		runes = append(runes, '━')
	}
	return string(runes)
}

// readStdin reads all of stdin. Returns empty string if stdin is a terminal.
func readStdin() (string, error) {
	info, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}
	if info.Mode()&os.ModeCharDevice != 0 {
		return "", nil // terminal, no piped input
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
