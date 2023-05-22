package model

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/ferama/yay/pkg/ai"
)

type nonInteractiveModel struct {
	spinner    spinner.Model
	requesting bool
	renderer   *glamour.TermRenderer

	request string

	ai  *ai.AI
	err error

	out string
}

type doReqMsg struct {
	Content string
}

func NewNonInteractiveModel(req string) *nonInteractiveModel {

	s := spinner.New()
	s.Spinner = spinner.Points

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithPreservedNewLines(),
	)

	return &nonInteractiveModel{
		spinner:    s,
		requesting: false,
		renderer:   renderer,
		request:    req,

		ai:  ai.NewAI(),
		err: nil,
	}
}

func (m *nonInteractiveModel) Output() string {
	var ret string
	if m.err == nil {
		ret, _ = m.renderer.Render(m.out)
	} else {
		if m.err == ai.ErrInvalidApiKey {
			errorMsg := "Api key is not valid. Is the 'OPENAI_API_KEY' env var set?\n"
			errorMsg += "You can grab one at https://platform.openai.com/account/api-keys\n"
			ret = errorMsg
		} else {
			ret = errorSytle.Render(fmt.Sprintf("[ERROR: %s]", m.err))
		}

	}
	return ret
}

func (m *nonInteractiveModel) Init() tea.Cmd {

	req := func() tea.Msg {
		return doReqMsg{
			Content: m.request,
		}
	}
	return tea.Batch(m.spinner.Tick, req)
}

func (m *nonInteractiveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case aiResponsesMsg:

		m.out = msg.Content
		return m, tea.Quit

	case doReqMsg:
		m.requesting = true
		cmd = func() tea.Msg {
			res, err := m.ai.SendMsg(msg.Content)
			if err != nil {
				return err
			}
			return aiResponsesMsg{
				Content: res,
			}
		}
		cmds = append(cmds, cmd)

	// We handle errors just like any other message
	case errMsg:
		m.requesting = false
		m.err = msg
		return m, tea.Quit
	}

	m.spinner, cmd = m.spinner.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *nonInteractiveModel) View() string {
	spin := "⎮ "
	if m.requesting {
		spin = m.spinner.View()
	}

	s := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return fmt.Sprintf(
		s.Render("Elaborating %s"),
		spin,
	)
}
