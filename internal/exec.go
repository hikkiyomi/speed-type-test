package internal

import (
	"fmt"
	"log"
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
	wrapWords      int
}

func NewModel(quote string, timeout int, wrapWords int) model {
	statuses := make([]status, len(quote))

	if len(quote) == 0 {
		log.Fatal("There should be at least one character in test")
	}

	if timeout < 0 {
		log.Fatal("Timeout should be at least zero")
	}

	if wrapWords <= 0 {
		log.Fatal("The number of words in one line should be more than zero")
	}

	return model{
		quote:          quote,
		statuses:       statuses,
		initialTimeout: timeout,
		timer:          timer.NewWithInterval(time.Duration(timeout)*time.Second, time.Millisecond),
		keymap: keymap{
			start: key.NewBinding(
				key.WithKeys(string(quote[0])),
			),
			end: key.NewBinding(
				key.WithKeys("ctrl+c"),
				key.WithHelp("ctrl+c", "end"),
			),
			quit: key.NewBinding(
				key.WithKeys("ctrl+c"),
				key.WithHelp("ctrl+c", "quit the program (when timer is stopped)"),
			),
		},
		help:      help.New(),
		wrapWords: wrapWords,
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

type testEndingMsg int

func (m *model) acceptInput(input string) tea.Cmd {
	if !checkInput(input) {
		return nil
	}

	inputCharacter := input[0]

	if m.quote[m.current] == inputCharacter {
		m.statuses[m.current] = STATUS_CORRECT
	} else {
		m.statuses[m.current] = STATUS_WRONG
	}

	m.current++

	for m.current < len(m.quote) && m.quote[m.current] == ' ' {
		m.current++
	}

	if m.current == len(m.quote) {
		return func() tea.Msg {
			return testEndingMsg(1)
		}
	}

	return nil
}

func (m model) testStarted() bool {
	return m.timer.Timeout != time.Duration(m.initialTimeout)*time.Second
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
	case timer.TimeoutMsg, testEndingMsg:
		m.isTyping = false
		return m, m.timer.Stop()
	case tea.KeyMsg:
		input := msg.String()

		switch {
		case key.Matches(msg, m.keymap.quit) && !m.isTyping:
			return m, tea.Quit
		case key.Matches(msg, m.keymap.start) && !m.testStarted():
			m.isTyping = true
			cmd := m.acceptInput(input)
			return m, tea.Batch(m.timer.Init(), cmd)
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

			cmd := m.acceptInput(input)
			return m, cmd
		}
	}

	return m, nil
}

func (m model) helpView() string {
	return "\n" + m.help.ShortHelpView([]key.Binding{
		m.keymap.end,
		m.keymap.quit,
	})
}

type testStats struct {
	correctWords int
	correctChars int
}

func (m model) getTestStats() testStats {
	correctWords := func() int {
		result := 0

		for i := 0; i < len(m.quote); i++ {
			if m.quote[i] == ' ' {
				continue
			}

			j := i
			isCorrect := true

			for j+1 < len(m.quote) && m.quote[j] != ' ' {
				isCorrect = isCorrect && (m.statuses[j] == STATUS_CORRECT)
				j++
			}

			if isCorrect {
				result++
			}

			i = j
		}

		return result
	}()

	correctChars := func() int {
		result := 0

		for i, c := range m.quote {
			if c == ' ' {
				continue
			}

			if m.statuses[i] == STATUS_CORRECT {
				result++
			}
		}

		return result
	}()

	return testStats{correctWords: correctWords, correctChars: correctChars}
}

func (m model) getTestResult() string {
	testStats := m.getTestStats()
	timePassed := m.initialTimeout - int(m.timer.Timeout/time.Second)
	avgWpm := 0
	avgCpm := 0

	if m.testStarted() {
		avgWpm = int(float64(testStats.correctWords) / float64(timePassed) * 60)
		avgCpm = int(float64(testStats.correctChars) / float64(timePassed) * 60)
	}

	return fmt.Sprintf("%v wpm, %v cpm", avgWpm, avgCpm)
}

func (m model) View() string {
	styledRunes := make([][]string, 0)
	currentRow := make([]string, 0)
	countWords := 0

	for i, c := range m.quote {
		style := styleMapping[m.statuses[i]]

		if m.current == i {
			style = style.Underline(true)
		}

		currentRow = append(currentRow, style.Render(string(c)))

		if c == ' ' {
			countWords++

			if countWords == m.wrapWords {
				styledRunes = append(styledRunes, currentRow)
				currentRow = make([]string, 0)
				countWords = 0
			}
		}
	}

	if len(currentRow) > 0 {
		styledRunes = append(styledRunes, currentRow)
	}

	header := func() string {
		addition := ""

		if !m.testStarted() {
			addition = " - start typing to start the test"
		}

		return lipgloss.JoinHorizontal(
			lipgloss.Right,
			m.timer.View(),
			addition,
		)
	}()

	mainPart := func() string {
		result := ""

		for _, line := range styledRunes {
			result = lipgloss.JoinVertical(
				lipgloss.Top,
				result,
				lipgloss.JoinHorizontal(
					lipgloss.Left,
					line...,
				),
			)
		}

		return result
	}()

	footer := lipgloss.JoinVertical(
		lipgloss.Top,
		"",
		m.getTestResult(),
		m.helpView(),
	)

	return lipgloss.Place(
		m.sizes.width,
		m.sizes.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			mainPart,
			footer,
		),
	)
}
