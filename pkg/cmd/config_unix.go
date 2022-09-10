//go:build !windows
// +build !windows

package cmd

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type passwordTextInputModel struct {
	textInput textinput.Model
}

func newPasswordTextInputModel(prompt string) passwordTextInputModel {
	textInput := textinput.New()
	textInput.Prompt = prompt
	textInput.Placeholder = "password"
	textInput.EchoMode = textinput.EchoNone
	textInput.Focus()
	return passwordTextInputModel{
		textInput: textInput,
	}
}

func (m passwordTextInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m passwordTextInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m passwordTextInputModel) Value() string {
	return m.textInput.Value()
}

func (m passwordTextInputModel) View() string {
	return m.textInput.View()
}

// readPassword reads a password.
func (c *Config) readPassword(prompt string) (password string, err error) {
	if c.noTTY {
		password, err = c.readLine(prompt)
		return
	}

	if c.PINEntry.Command != "" {
		return c.readPINEntry(prompt)
	}

	program := tea.NewProgram(newPasswordTextInputModel(prompt))
	model, err := program.StartReturningModel()
	if err != nil {
		return "", err
	}
	//nolint:forcetypeassert
	value := model.(passwordTextInputModel).Value()
	return value, nil
}

func (c *Config) windowsVersion() (map[string]any, error) {
	return nil, nil
}
