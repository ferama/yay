package model

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/ferama/yay/pkg/ai"
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

	errorSytle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9"))

	youMSG = youStyle.Render("You")
	aiMSG  = aiStyle.Render("AI")
)

type aiResponsesMsg struct {
	Content string
}

type (
	errMsg error
)

type interactiveModel struct {
	textInput  textinput.Model
	spinner    spinner.Model
	requesting bool
	renderer   *glamour.TermRenderer

	ai  *ai.AI
	err error
}

func NewInteractiveModel() *interactiveModel {

	ti := textinput.New()
	ti.Placeholder = "Send a msg"
	ti.Focus()
	ti.CharLimit = 1000
	ti.Width = 60

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithPreservedNewLines(),
	)

	return &interactiveModel{
		spinner:    s,
		textInput:  ti,
		requesting: false,
		renderer:   renderer,

		ai:  ai.NewAI(),
		err: nil,
	}
}

func (m *interactiveModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m *interactiveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case aiResponsesMsg:
		out, _ := m.renderer.Render(msg.Content)

		cmds = append(cmds, tea.Printf("%s%s", aiMSG, out))
		m.requesting = false

	case tea.WindowSizeMsg:
		m.textInput.Width = msg.Width

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
			cmds = append(cmds, tea.Printf("%s\n  %s\n", youMSG, value))

			m.requesting = true
			cmd = func() tea.Msg {
				res, err := m.ai.SendMsg(value)
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
		m.requesting = false
		m.err = msg
		if m.err == ai.ErrInvalidApiKey {
			m.View()
			return m, tea.Quit
		}
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *interactiveModel) View() string {
	spin := "⎮ "
	if m.requesting {
		spin = m.spinner.View()
	}

	err := ""
	if m.err != nil {
		if m.err == ai.ErrInvalidApiKey {
			err = "\nApi key is not valid. Is the 'OPENAI_API_KEY' env var set?\n"
			err += "You can grab one at https://platform.openai.com/account/api-keys\n\n"
		} else {
			err = fmt.Sprintf("[ERROR: %s]", m.err)
		}
	}

	return fmt.Sprintf(
		"\n%s%s\n%s %s",
		spin,
		m.textInput.View(),
		"(esc to quit)",
		errorSytle.Render(err),
	)
}
