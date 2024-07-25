package internal

import (
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/stopwatch"
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
	quote      string
	current    int
	isTyping   bool
	statuses   []status
	timeModule TimeModule
	keymap     keymap
	help       help.Model
	sizes      windowSizes
	wrapWords  int
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

	var timeModule TimeModule

	if timeout == 0 {
		timeModule = newStopwatch(stopwatch.NewWithInterval(time.Millisecond))
	} else {
		timeModule = newTimer(
			timer.NewWithInterval(
				time.Duration(timeout)*time.Second, time.Millisecond,
			),
			timeout,
		)
	}

	return model{
		quote:      quote,
		statuses:   statuses,
		timeModule: timeModule,
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

		for m.current > 0 && m.quote[m.current] == ' ' {
			m.current--
			m.statuses[m.current] = STATUS_PENDING
		}
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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.sizes.width = msg.Width
		m.sizes.height = msg.Height
	case timer.TimeoutMsg, testEndingMsg:
		m.isTyping = false
		return m, m.timeModule.Stop()
	case tea.KeyMsg:
		input := msg.String()

		switch {
		case key.Matches(msg, m.keymap.quit) && !m.isTyping:
			return m, tea.Quit
		case key.Matches(msg, m.keymap.start) && !m.timeModule.HasStarted():
			m.isTyping = true
			cmd := m.acceptInput(input)
			return m, tea.Batch(m.timeModule.Init(), cmd)
		case key.Matches(msg, m.keymap.end):
			m.isTyping = false
			return m, m.timeModule.Stop()
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

	var cmd tea.Cmd
	m.timeModule, cmd = m.timeModule.Update(msg)

	return m, cmd
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
	timePassed := m.timeModule.GetTimePassed()
	avgWpm := 0
	avgCpm := 0

	if timePassed > 0 {
		avgWpm = int(float64(testStats.correctWords) / float64(timePassed) * 60)
		avgCpm = int(float64(testStats.correctChars) / float64(timePassed) * 60)
	}

	return fmt.Sprintf("%v wpm, %v cpm", avgWpm, avgCpm)
}

func getStyledRows(m model) ([][]string, *int) {
	styledRunes := make([][]string, 0)
	currentRow := make([]string, 0)
	countWords := 0

	var rowWithCursor *int

	for i, c := range m.quote {
		style := styleMapping[m.statuses[i]]

		if m.current == i {
			style = style.Underline(true)
			rowWithCursor = new(int)
			*rowWithCursor = len(styledRunes)
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

	return styledRunes, rowWithCursor
}

func getRenderingRows(m model) [][]string {
	styledRunes, rowWithCursor := getStyledRows(m)
	rowsToRender := make([][]string, 0)

	if rowWithCursor != nil {
		row := *rowWithCursor

		if row > 0 {
			rowsToRender = append(rowsToRender, styledRunes[row-1])
		}

		rowsToRender = append(rowsToRender, styledRunes[row])

		if row+1 < len(styledRunes) {
			rowsToRender = append(rowsToRender, styledRunes[row+1])
		}
	} else {
		for i := 3; i >= 1; i-- {
			if len(styledRunes)-i >= 0 {
				rowsToRender = append(rowsToRender, styledRunes[len(styledRunes)-i])
			}
		}
	}

	return rowsToRender
}

func (m model) View() string {
	rowsToRender := getRenderingRows(m)

	header := func() string {
		addition := ""

		if !m.timeModule.HasStarted() {
			addition = " - start typing to start the test"
		}

		return lipgloss.JoinHorizontal(
			lipgloss.Right,
			m.timeModule.View(),
			addition,
		)
	}()

	mainPart := func() string {
		result := ""

		for _, line := range rowsToRender {
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
