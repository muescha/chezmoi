//go:build !windows
// +build !windows

package cmd

import (
	"fmt"

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
	return passwordTextInputModel{
		textInput: textInput,
	}
}

func (m passwordTextInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m passwordTextInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	fmt.Printf("msg=%+v\n", msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	fmt.Printf("m=%+v\n", m)
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

	m := newPasswordTextInputModel(prompt)
	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		return "", err
	}
	fmt.Printf("value=%q\n", m.Value())
	return m.Value(), nil
}

func (c *Config) windowsVersion() (map[string]any, error) {
	return nil, nil
}
