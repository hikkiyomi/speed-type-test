package internal

import (
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type status int

const (
	STATUS_PENDING status = 0
	STATUS_CORRECT status = 1
	STATUS_WRONG   status = 2
)

type keymap struct {
	start key.Binding
	end   key.Binding
	quit  key.Binding
}

type windowSizes struct {
	width  int
	height int
}

type model struct {
	quote          string
	current        int
	isTyping       bool
	statuses       []status
	initialTimeout int
	timer          timer.Model
	keymap         keymap
	help           help.Model
	sizes          windowSizes
}

func NewModel(quote string, timeout int) model {
	statuses := make([]status, len(quote))

	return model{
		quote:          quote,
		statuses:       statuses,
		initialTimeout: timeout,
		timer:          timer.NewWithInterval(time.Duration(timeout)*time.Second, time.Millisecond),
		keymap: keymap{
			start: key.NewBinding(
				key.WithKeys("enter"),
				key.WithHelp("enter", "start"),
			),
			end: key.NewBinding(
				key.WithKeys("ctrl+c"),
				key.WithHelp("ctrl+c", "end"),
			),
			quit: key.NewBinding(
				key.WithKeys("q"),
				key.WithHelp("q", "quit (works only when timer is stopped)"),
			),
		},
		help: help.New(),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) applyBackspace() {
	if m.current > 0 {
		m.current--
		m.statuses[m.current] = STATUS_PENDING
	}
}

func checkInput(input string) bool {
	return len(input) == 1 &&
		(('a' <= input[0] && input[0] <= 'z') ||
			('A' <= input[0] && input[0] <= 'Z') ||
			input[0] == ' ')
}

func (m *model) acceptInput(input string) {
	if !checkInput(input) {
		return
	}

	inputCharacter := input[0]

	if m.quote[m.current] == inputCharacter {
		m.statuses[m.current] = STATUS_CORRECT
	} else {
		m.statuses[m.current] = STATUS_WRONG
	}

	m.current++
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.sizes.width = msg.Width
		m.sizes.height = msg.Height
	case timer.TickMsg, timer.StartStopMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd
	case timer.TimeoutMsg:
		m.isTyping = false
	case tea.KeyMsg:
		input := msg.String()

		switch {
		case key.Matches(msg, m.keymap.quit) && !m.isTyping:
			return m, tea.Quit
		case key.Matches(msg, m.keymap.start) && m.timer.Timeout == time.Duration(m.initialTimeout)*time.Second:
			m.isTyping = true
			return m, m.timer.Init()
		case key.Matches(msg, m.keymap.end):
			m.isTyping = false
			return m, m.timer.Stop()
		case input == "backspace":
			if !m.isTyping {
				break
			}

			m.applyBackspace()
		default:
			if !m.isTyping {
				break
			}

			m.acceptInput(input)
		}
	}

	return m, nil
}

func (m model) helpView() string {
	return "\n" + m.help.ShortHelpView([]key.Binding{
		m.keymap.start,
		m.keymap.end,
		m.keymap.quit,
	})
}

func (m model) View() string {
	styledRunes := make([]string, 0)

	for i, c := range m.quote {
		style := styleMapping[m.statuses[i]]

		if m.current == i {
			style = style.Underline(true)
		}

		styledRunes = append(styledRunes, style.Render(string(c)))
	}

	return lipgloss.Place(
		m.sizes.width,
		m.sizes.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Left,
			m.timer.View(),
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				styledRunes...,
			),
			m.helpView(),
		),
	)
}
