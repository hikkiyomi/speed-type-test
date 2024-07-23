package internal

import (
	"time"

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

type model struct {
	quote    string
	current  int
	isTyping bool
	statuses []status
	timer    timer.Model
}

func NewModel(quote string, timeout int) model {
	statuses := make([]status, len(quote))

	return model{
		quote:    quote,
		statuses: statuses,
		timer:    timer.NewWithInterval(time.Duration(timeout)*time.Second, time.Millisecond),
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
	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)

		return m, cmd
	case timer.TimeoutMsg:
		return m, tea.Quit
	case tea.KeyMsg:
		input := msg.String()

		switch input {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			if !m.isTyping {
				m.isTyping = true
			}

			return m, m.timer.Init()
		case "backspace":
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

func (m model) View() string {
	styledRunes := make([]string, 0)

	for i, c := range m.quote {
		style := styleMapping[m.statuses[i]]

		if m.current == i {
			style = style.Underline(true)
		}

		styledRunes = append(styledRunes, style.Render(string(c)))
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.timer.View(),
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			styledRunes...,
		),
	)
}
