package internal

import (
	"time"

	"github.com/charmbracelet/bubbles/stopwatch"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
)

type TimeModule interface {
	Init() tea.Cmd
	Update(tea.Msg) (TimeModule, tea.Cmd)
	View() string
	HasStarted() bool
	Stop() tea.Cmd
	GetTimePassed() int
}

type Timer struct {
	timer          timer.Model
	initialTimeout int
}

type Stopwatch struct {
	stopwatch stopwatch.Model
}

// TIMER PART

func newTimer(model timer.Model, initialTimeout int) Timer {
	return Timer{timer: model, initialTimeout: initialTimeout}
}

func (t Timer) Init() tea.Cmd {
	return t.timer.Init()
}

func (t Timer) Update(msg tea.Msg) (TimeModule, tea.Cmd) {
	var cmd tea.Cmd
	t.timer, cmd = t.timer.Update(msg)
	return t, cmd
}

func (t Timer) View() string {
	return t.timer.View()
}

func (t Timer) HasStarted() bool {
	return t.timer.Timeout != time.Duration(t.initialTimeout)*time.Second
}

func (t Timer) Stop() tea.Cmd {
	return t.timer.Stop()
}

func (t Timer) GetTimePassed() int {
	return t.initialTimeout - int(t.timer.Timeout/time.Second)
}

// STOPWATCH PART

func newStopwatch(model stopwatch.Model) Stopwatch {
	return Stopwatch{model}
}

func (s Stopwatch) Init() tea.Cmd {
	return s.stopwatch.Init()
}

func (s Stopwatch) Update(msg tea.Msg) (TimeModule, tea.Cmd) {
	var cmd tea.Cmd
	s.stopwatch, cmd = s.stopwatch.Update(msg)
	return s, cmd
}

func (s Stopwatch) View() string {
	return s.stopwatch.View()
}

func (s Stopwatch) HasStarted() bool {
	return s.stopwatch.Elapsed().Nanoseconds() > 0
}

func (s Stopwatch) Stop() tea.Cmd {
	return s.stopwatch.Stop()
}

func (s Stopwatch) GetTimePassed() int {
	return int(s.stopwatch.Elapsed() / time.Second)
}
