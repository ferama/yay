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
	yayCmd.Flags().StringP("header", "e", "", "add message header")
}

func yayPipe(header string) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if stat.Mode()&os.ModeNamedPipe == 0 && stat.Size() == 0 {
		fmt.Println("Try piping in some text.")
		os.Exit(1)
	}
	reader := &io.LimitedReader{
		R: bufio.NewReader(os.Stdin),
		N: 1024 * 10,
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		fmt.Printf("Error while reading from pipe: %s", err.Error())
		os.Exit(1)
	}

	var msg string
	msg = string(body)
	if header != "" {
		msg = fmt.Sprintf("%s\n%s", header, body)
	}

	model := model.NewNonInteractiveModel(msg)
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	fmt.Println(model.Output())
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
			return
		}

		header, _ := cmd.Flags().GetString("header")
		yayPipe(header)
	},
}

// Execute executes the yay command
func Execute() error {
	return yayCmd.Execute()
}
