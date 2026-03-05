package cli

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:          "kno",
		Short:        "Local-first knowledge capture and workflow tool",
		Long:         "kno captures valuable LLM sessions into a local knowledge vault.",
		SilenceUsage: true,
	}

	root.AddCommand(
		newSetupCmd(),
		newCaptureCmd(),
		newListCmd(),
		newShowCmd(),
		newMCPCmd(),
		newVersionCmd(),
	)

	return root
}
