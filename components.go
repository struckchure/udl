package udl

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	padding  = 2
	maxWidth = 80
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

// ProgressBar is a reusable TUI progress bar.
type ProgressBar struct {
	model model
	prog  *tea.Program
}

// model is the Bubbletea model for the progress bar.
type model struct {
	percent  float64
	progress progress.Model
}

// progressMsg is sent to update the progress bar from outside.
type progressMsg float64

// NewProgressBar creates a new reusable progress bar.
func NewProgressBar() *ProgressBar {
	prog := progress.New(progress.WithScaledGradient("#FF7CCB", "#FDFF8C"))
	m := model{progress: prog}
	return &ProgressBar{
		model: m,
		prog:  tea.NewProgram(m),
	}
}

// Start runs the progress bar in a goroutine.
func (pb *ProgressBar) Start() {
	go func() {
		if _, err := pb.prog.Run(); err != nil {
			fmt.Println("Progress bar error:", err)
			os.Exit(1)
		}
	}()
}

// Update sets the progress to a given percent (0.0 - 1.0).
func (pb *ProgressBar) Update(percent float64) {
	pb.prog.Send(progressMsg(percent))
}

// Stop stops the progress bar program.
func (pb *ProgressBar) Stop() {
	pb.prog.Quit()
}

// ---- Bubbletea implementation ----

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - padding*2 - 4
		if m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
		}
		return m, nil

	case progressMsg:
		if msg < 0 {
			msg = 0
		}
		if msg > 1 {
			msg = 1
		}
		m.percent = float64(msg)
		if m.percent == 1 {
			return m, tea.Quit
		}
		return m, nil

	default:
		return m, nil
	}
}

func (m model) View() string {
	pad := strings.Repeat(" ", padding)
	return "\n" +
		pad + m.progress.ViewAs(m.percent) + "\n\n" +
		pad + helpStyle("Press any key to quit")
}
