package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/kno-ai/kno/internal/app"
	"github.com/kno-ai/kno/internal/capture"
	"github.com/kno-ai/kno/internal/config"
	"github.com/spf13/cobra"
)

func newCaptureCmd() *cobra.Command {
	var (
		clipboard bool
		stdin     bool
		title     string
		metaPairs []string
	)

	cmd := &cobra.Command{
		Use:   "capture [file]",
		Short: "Capture content into the knowledge vault",
		Long:  "Capture content from clipboard, stdin, or a file into the vault.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Determine source.
			sourceCount := 0
			if clipboard {
				sourceCount++
			}
			if stdin {
				sourceCount++
			}
			if len(args) > 0 {
				sourceCount++
			}
			if sourceCount != 1 {
				return fmt.Errorf("exactly one source required: --clipboard, --stdin, or a file path")
			}

			var body string
			var sourceKind string

			switch {
			case clipboard:
				content, err := readClipboard()
				if err != nil {
					return fmt.Errorf("reading clipboard: %w", err)
				}
				body = content
				sourceKind = "clipboard"

			case stdin:
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("reading stdin: %w", err)
				}
				body = string(data)
				sourceKind = "stdin"

			default:
				data, err := os.ReadFile(args[0])
				if err != nil {
					return fmt.Errorf("reading file: %w", err)
				}
				body = string(data)
				sourceKind = "file"
			}

			if body == "" {
				return fmt.Errorf("no content to capture")
			}

			meta, err := parseMeta(metaPairs)
			if err != nil {
				return err
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config (have you run 'kno setup'?): %w", err)
			}

			a, err := app.FromConfig(cfg)
			if err != nil {
				return err
			}

			result, err := a.Capture.Create(capture.CreateParams{
				Title:      title,
				BodyMD:     body,
				SourceKind: sourceKind,
				SourceTool: "kno_cli",
				Meta:       meta,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", result.Path)
			return nil
		},
	}

	cmd.Flags().BoolVar(&clipboard, "clipboard", false, "Read from system clipboard")
	cmd.Flags().BoolVar(&stdin, "stdin", false, "Read from stdin")
	cmd.Flags().StringVar(&title, "title", "", "Title for the capture")
	cmd.Flags().StringArrayVar(&metaPairs, "meta", nil, "Metadata key=value pair (repeatable)")

	return cmd
}

func parseMeta(pairs []string) (map[string]string, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	m := make(map[string]string, len(pairs))
	for _, p := range pairs {
		k, v, ok := strings.Cut(p, "=")
		if !ok {
			return nil, fmt.Errorf("invalid --meta value %q: expected key=value", p)
		}
		m[k] = v
	}
	return m, nil
}

func readClipboard() (string, error) {
	if runtime.GOOS != "darwin" {
		return "", fmt.Errorf("clipboard capture is only supported on macOS")
	}
	out, err := exec.Command("pbpaste").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
