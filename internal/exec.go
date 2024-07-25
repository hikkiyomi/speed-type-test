package internal

import (
	"fmt"
	"log"
	"strings"
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

type testStats struct {
	correctWords int
	correctChars int
}

type model struct {
	quote      *quote
	isTyping   bool
	timeModule TimeModule
	keymap     keymap
	help       help.Model
	sizes      windowSizes
	wrapWords  int
	stats      testStats
}

func NewModel(quote string, timeout int, wrapWords int) model {
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
		quote:      newQuote(quote, wrapWords),
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
	m.quote.Prev()

	if m.quote.GetCurrentChar().Status == STATUS_CORRECT {
		m.stats.correctChars--
	}

	isWordCorrect := true

	for _, c := range m.quote.Words[m.quote.Row][m.quote.WordPos] {
		if c.Status != STATUS_CORRECT {
			isWordCorrect = false
			break
		}
	}

	if isWordCorrect {
		m.stats.correctWords--
	}

	m.quote.GetCurrentChar().Status = STATUS_PENDING
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

	if m.quote.GetCurrentChar().Value == inputCharacter {
		m.quote.GetCurrentChar().Status = STATUS_CORRECT

		m.stats.correctChars++
		isWordCorrect := true

		for _, c := range m.quote.Words[m.quote.Row][m.quote.WordPos] {
			if c.Status != STATUS_CORRECT {
				isWordCorrect = false
				break
			}
		}

		if isWordCorrect {
			m.stats.correctWords++
		}
	} else {
		m.quote.GetCurrentChar().Status = STATUS_WRONG
	}

	canMove := m.quote.Next()

	if !canMove {
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

func (m model) getTestResult() string {
	testStats := m.stats
	timePassed := m.timeModule.GetTimePassed()
	avgWpm := 0
	avgCpm := 0

	if timePassed > 0 {
		avgWpm = int(float64(testStats.correctWords) / float64(timePassed) * 60)
		avgCpm = int(float64(testStats.correctChars) / float64(timePassed) * 60)
	}

	return fmt.Sprintf("%v wpm, %v cpm", avgWpm, avgCpm)
}

func getRenderingRows(m model) []string {
	rowsToRender := make([]string, 3)

	renderRow := func(row int, underlinedWord *int, underlinedChar *int) string {
		renderedWords := make([]string, 0)

		for i, w := range m.quote.Words[row] {
			if underlinedWord != nil && i == *underlinedWord {
				renderedWords = append(renderedWords, w.Render(underlinedChar))
			} else {
				renderedWords = append(renderedWords, w.Render(nil))
			}
		}

		return strings.Join(renderedWords, " ")
	}

	if m.quote.Row == 0 {
		rowsToRender[0] = renderRow(0, &m.quote.WordPos, &m.quote.Pos)

		for i := 1; i <= 2; i++ {
			if m.quote.Row+i < len(m.quote.Words) {
				rowsToRender[i] = renderRow(i, nil, nil)
			}
		}
	} else {
		rowsToRender[1] = renderRow(m.quote.Row, &m.quote.WordPos, &m.quote.Pos)

		for i := -1; i <= 1; i += 2 {
			if m.quote.Row+i < len(m.quote.Words) {
				rowsToRender[1+i] = renderRow(m.quote.Row+i, nil, nil)
			}
		}
	}

	return rowsToRender
}

func (m model) View() string {
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
		rowsToRender := getRenderingRows(m)

		return lipgloss.JoinVertical(
			lipgloss.Top,
			rowsToRender...,
		)
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
