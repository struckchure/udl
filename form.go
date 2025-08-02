package udl

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
)

type InputModel struct {
	Input textinput.Model
	err   error
}

func InputModelForm() InputModel {
	ti := textinput.New()
	ti.Placeholder = "tom and jerry"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return InputModel{Input: ti, err: nil}
}

func (m InputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m InputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	// We handle errors just like any other message
	case error:
		m.err = msg
		return m, nil
	}

	m.Input, cmd = m.Input.Update(msg)

	return m, cmd
}

func (m InputModel) View() string {
	return fmt.Sprintf("Search \n\n%s\n\n%s", m.Input.View(), "(esc to quit)") + "\n"
}

// select form

type SelectModel struct {
	Title, Value, Desc string
}

func (i SelectModel) FilterValue() string { return i.Title }

type SelectModelForm struct {
	Model list.Model
}

func (m SelectModelForm) Init() tea.Cmd {
	return nil
}

func (m SelectModelForm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
		case "enter":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.Model.SetSize(msg.Width, msg.Height)
	}

	var cmd tea.Cmd
	m.Model, cmd = m.Model.Update(msg)
	return m, cmd
}

func (m SelectModelForm) View() string {
	return m.Model.View()
}

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
)

type ItemDelegate struct{}

func (d ItemDelegate) Height() int                             { return 1 }
func (d ItemDelegate) Spacing() int                            { return 0 }
func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(SelectModel)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i.Title)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " ") + lo.Ternary(i.Desc != "", " / "+i.Desc, ""))
		}
	}

	fmt.Fprint(w, fn(str))
}
