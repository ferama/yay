package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ferama/yay/pkg/model"
	"github.com/spf13/cobra"
)

func init() {
	yayCmd.Flags().BoolP("interactive", "i", false, "use interactive mode")
}

func yayPipe(header string) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	pipedContent := true
	if stat.Mode()&os.ModeNamedPipe == 0 && stat.Size() == 0 {
		pipedContent = false
	}

	var msg string

	if pipedContent {
		reader := bufio.NewReader(os.Stdin)
		body, err := io.ReadAll(reader)
		if err != nil {
			fmt.Printf("Error while reading from pipe: %s", err.Error())
			os.Exit(1)
		}

		msg = string(body)
		if header != "" {
			msg = fmt.Sprintf("%s\n%s", header, body)
		}
	} else {
		if header != "" {
			msg = header
		} else {
			fmt.Println("no piped content and no header is present")
			os.Exit(1)
		}
	}

	model := model.NewNonInteractiveModel(msg)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(model.Output())
}

var yayCmd = &cobra.Command{
	Use: "yay",
	Args: func(cmd *cobra.Command, args []string) error {
		isInteractive, _ := cmd.Flags().GetBool("interactive")
		if !isInteractive {
			if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
				return err
			}
		}
		return nil
	},
	Example: `
  # starts yay in interactive mode
  $ yay -i
  
  # pipe in input
  $ ls ~/tmp | yay "could you order this files by type?"
`,
	Run: func(cmd *cobra.Command, args []string) {
		isInteractive, _ := cmd.Flags().GetBool("interactive")

		if isInteractive {
			p := tea.NewProgram(model.NewInteractiveModel())
			if _, err := p.Run(); err != nil {
				log.Fatal(err)
			}
			return
		}

		header := os.Args[1]
		yayPipe(header)
	},
}

// Execute executes the yay command
func Execute() error {
	return yayCmd.Execute()
}
