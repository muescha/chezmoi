package cmd

import (
	"bufio"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/twpayne/chezmoi/v2/pkg/chezmoi"
)

type passwordTextInputModel struct {
	textInput textinput.Model
	aborted   bool
}

type stringInputModel struct {
	textInput textinput.Model
	aborted   bool
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

func (m passwordTextInputModel) Aborted() bool {
	return m.aborted
}

func (m passwordTextInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m passwordTextInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	//nolint:gocritic
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.aborted = true
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

func newStringInputModel(prompt string) stringInputModel {
	textInput := textinput.New()
	textInput.Prompt = prompt
	textInput.Focus()
	return stringInputModel{
		textInput: textInput,
	}
}

func (m stringInputModel) Aborted() bool {
	return m.aborted
}

func (m stringInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m stringInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	//nolint:gocritic
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.aborted = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m stringInputModel) Value() string {
	return m.textInput.Value()
}

func (m stringInputModel) View() string {
	return m.textInput.View()
}

// readLineRaw reads a line from stdin, trimming leading and trailing
// whitespace.
func (c *Config) readLineRaw(prompt string) (string, error) {
	_, err := c.stdout.Write([]byte(prompt))
	if err != nil {
		return "", err
	}
	if c.bufioReader == nil {
		c.bufioReader = bufio.NewReader(c.stdin)
	}
	line, err := c.bufioReader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// readPassword reads a password.
func (c *Config) readPassword(prompt string) (string, error) {
	if c.noTTY {
		return c.readLineRaw(prompt)
	}

	if c.PINEntry.Command != "" {
		return c.readPINEntry(prompt)
	}

	initModel := newPasswordTextInputModel(prompt)
	program := tea.NewProgram(initModel)
	finalModel, err := program.StartReturningModel()
	if err != nil {
		return "", err
	}
	//nolint:forcetypeassert
	model := finalModel.(passwordTextInputModel)
	if model.Aborted() {
		return "", chezmoi.ExitCodeError(1)
	}
	return model.Value(), nil
}

// readString reads a string.
func (c *Config) readString(prompt string) (string, error) {
	if c.noTTY {
		return c.readLineRaw(prompt)
	}

	initModel := newStringInputModel(prompt)
	program := tea.NewProgram(initModel)
	finalModel, err := program.StartReturningModel()
	if err != nil {
		return "", err
	}
	//nolint:forcetypeassert
	model := finalModel.(stringInputModel)
	if model.Aborted() {
		return "", chezmoi.ExitCodeError(1)
	}
	return model.Value(), nil
}
