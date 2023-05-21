package main

// A simple program demonstrating the text input component from the Bubbles
// component library.

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

var (
	youStyle = lipgloss.NewStyle().
			Bold(true).
			Underline(true).
			Foreground(lipgloss.Color("5"))

	aiStyle = lipgloss.NewStyle().
		Bold(true).
		Underline(true).
		Foreground(lipgloss.Color("202"))

	youMSG = youStyle.Render("You")
	aiMSG  = aiStyle.Render("AI")
)

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type aiResponsesMsg struct {
	Content string
}

type (
	errMsg error
)

type model struct {
	textInput  textinput.Model
	spinner    spinner.Model
	requesting bool
	renderer   *glamour.TermRenderer

	ai  *AI
	err error
}

func initialModel() model {

	ti := textinput.New()
	ti.Placeholder = "Send a msg"
	ti.Focus()
	ti.CharLimit = 1000
	ti.Width = 60

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	renderer, _ := glamour.NewTermRenderer(glamour.WithAutoStyle())

	return model{
		spinner:    s,
		textInput:  ti,
		requesting: false,
		renderer:   renderer,

		ai:  newAI(),
		err: nil,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case aiResponsesMsg:
		out, _ := m.renderer.Render(msg.Content)

		cmds = append(cmds, tea.Printf("%s\t%s", aiMSG, out))
		m.requesting = false

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			value := m.textInput.Value()
			m.textInput.Reset()
			if strings.TrimSpace(value) == "" {
				break
			}
			cmds = append(cmds, tea.Printf("%s\n  %s", youMSG, value))

			m.requesting = true
			cmd = func() tea.Msg {
				res, err := m.ai.sendMsg(value)
				if err != nil {
					return err
				}
				return aiResponsesMsg{
					Content: res,
				}
			}
			cmds = append(cmds, cmd)
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	spin := "⎮ "
	if m.requesting {
		spin = m.spinner.View()
	}

	return fmt.Sprintf(
		"\n%s%s\n%s",
		spin,
		m.textInput.View(),
		"(esc to quit)",
	)
}
