package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kno-ai/kno/internal"
	"github.com/kno-ai/kno/internal/update"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "kno",
		Short:         "Local-first knowledge vault",
		Long:          "kno saves and organizes knowledge from LLM sessions into a local vault.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			if !vaultExists(cmd) {
				fmt.Fprintln(cmd.ErrOrStderr(), "\nGet started: run 'kno setup' to create your vault.")
			}
		},
	}

	root.PersistentFlags().String("vault", "", "Path to the vault directory (default: ~/kno)")

	root.AddCommand(
		newSetupCmd(),
		newNoteCmd(),
		newPageCmd(),
		newVaultCmd(),
		newPublishCmd(),
		newMCPCmd(),
		newVersionCmd(),
	)

	return root
}

func vaultExists(cmd *cobra.Command) bool {
	vaultPath, _ := cmd.Flags().GetString("vault")
	if vaultPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		vaultPath = filepath.Join(home, "kno")
	}
	_, err := os.Stat(filepath.Join(vaultPath, "config.toml"))
	return err == nil
}

// Execute runs the root command and handles error output formatting.
func Execute() {
	// Check for updates in the background (non-blocking, once per day).
	// Skip for machine-oriented invocations (mcp, --json).
	updateMsg := make(chan string, 1)
	isMachine := len(os.Args) > 1 && os.Args[1] == "mcp"
	go func() {
		if isMachine {
			updateMsg <- ""
			return
		}
		updateMsg <- update.Check(internal.Version)
	}()

	root := NewRootCommand()
	if err := root.Execute(); err != nil {
		// Check if --json flag was set anywhere in the command chain
		jsonFlag := false
		if cmd, _, findErr := root.Find(os.Args[1:]); findErr == nil {
			if f := cmd.Flags().Lookup("json"); f != nil {
				jsonFlag = f.Changed
			}
		}

		if jsonFlag {
			data, _ := json.Marshal(map[string]string{"error": err.Error()})
			fmt.Fprintln(os.Stderr, string(data))
		} else {
			fmt.Fprintf(os.Stderr, "\u2717 %s\n", err.Error())
		}
		os.Exit(1)
	}

	select {
	case msg := <-updateMsg:
		if msg != "" {
			fmt.Fprintln(os.Stderr, msg)
		}
	case <-time.After(500 * time.Millisecond):
		// Don't block exit waiting for a slow network.
	}
}
