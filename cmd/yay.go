package cmd

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ferama/yay/pkg/model"
	"github.com/spf13/cobra"
)

func init() {
	yayCmd.Flags().BoolP("interactive", "i", false, "use interactive mode")
}

var yayCmd = &cobra.Command{
	Use:  "yay",
	Args: cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		isInteractive, _ := cmd.Flags().GetBool("interactive")

		if isInteractive {
			p := tea.NewProgram(model.NewInteractiveModel())
			if _, err := p.Run(); err != nil {
				log.Fatal(err)
			}
		}
	},
}

// Execute executes the yay command
func Execute() error {
	return yayCmd.Execute()
}
