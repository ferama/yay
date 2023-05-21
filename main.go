package main

// A simple program demonstrating the text input component from the Bubbles
// component library.

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type (
	errMsg error
)

type model struct {
	textInput textinput.Model
	ai        *AI
	err       error
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Ask"
	ti.Focus()
	ti.CharLimit = 1000
	ti.Width = 60

	return model{
		textInput: ti,
		ai:        newAI(),
		err:       nil,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			value := m.textInput.Value()
			m.textInput.Reset()
			res, err := m.ai.ask(value)
			if err != nil {
				m.err = err
				return m, nil
			}
			return m, tea.Printf("You:\t%s\nAI:\t%s", value, res)
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return fmt.Sprintf(
		"\n%s\n\n%s",
		m.textInput.View(),
		"(esc to quit)",
	)
}
