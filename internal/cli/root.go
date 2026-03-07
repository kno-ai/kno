package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "kno",
		Short:         "Local-first knowledge vault",
		Long:          "kno saves and organizes knowledge from LLM sessions into a local vault.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().String("vault", "", "Path to the vault directory (default: ~/kno)")

	root.AddCommand(
		newSetupCmd(),
		newNoteCmd(),
		newPageCmd(),
		newVaultCmd(),
		newAdminCmd(),
		newMCPCmd(),
		newVersionCmd(),
	)

	return root
}

// Execute runs the root command and handles error output formatting.
func Execute() {
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
}
