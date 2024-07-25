package internal

import (
	"bufio"
	"log"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/exp/rand"
)

type char struct {
	Value  byte
	Status status
}

type word []char

func (w word) GetStyles() []lipgloss.Style {
	result := make([]lipgloss.Style, 0)

	for _, c := range w {
		style := styleMapping[c.Status]
		result = append(result, style)
	}

	return result
}

// A function to render word. Pass a pointer to position of underlined character.
// Use nil pointer for no underlining.
func (w word) Render(underline *int) string {
	styles := w.GetStyles()
	result := ""

	for i, s := range styles {
		if underline != nil && i == *underline {
			s = s.Underline(true)
		}

		result += s.Render(string(w[i].Value))
	}

	return result
}

func newWord(flatWord string) word {
	chars := make([]char, len(flatWord))

	for i := 0; i < len(flatWord); i++ {
		chars[i].Value = flatWord[i]
		chars[i].Status = STATUS_PENDING
	}

	return word(chars)
}

type quote struct {
	Words   [][]word
	Row     int
	WordPos int
	Pos     int
}

func newQuote(flatQuote string, wrapWords int) *quote {
	wordsInQuote := strings.Split(flatQuote, " ")

	words := make([][]word, 0)
	row := make([]word, 0)

	for _, flatWord := range wordsInQuote {
		row = append(row, newWord(flatWord))

		if len(row) == wrapWords {
			words = append(words, row)
			row = make([]word, 0)
		}
	}

	if len(row) > 0 {
		words = append(words, row)
	}

	return &quote{
		Words: words,
	}
}

// Moves a 'pointer' to character to the right.
// Returns false if there is nowhere to move, otherwise returns true.
func (q *quote) Next() bool {
	result := true

	if q.Pos+1 < len(q.Words[q.Row][q.WordPos]) {
		q.Pos++
	} else if q.WordPos+1 < len(q.Words[q.Row]) {
		q.WordPos++
		q.Pos = 0
	} else if q.Row+1 < len(q.Words) {
		q.Row++
		q.WordPos = 0
		q.Pos = 0
	} else {
		result = false
	}

	return result
}

// Moves a 'pointer' to character to the left.
// Returns false if there is nowhere to move, otherwise returns true.
func (q *quote) Prev() bool {
	result := true

	if q.Pos > 0 {
		q.Pos--
	} else if q.WordPos > 0 {
		q.WordPos--
		q.Pos = len(q.Words[q.Row][q.WordPos]) - 1
	} else if q.Row > 0 {
		q.Row--
		q.WordPos = len(q.Words[q.Row]) - 1
		q.Pos = len(q.Words[q.Row][q.WordPos]) - 1
	} else {
		result = false
	}

	return result
}

func (q *quote) GetCurrentChar() *char {
	return &q.Words[q.Row][q.WordPos][q.Pos]
}

func shuffle(words []string) {
	rand.Seed(uint64(time.Now().UnixNano()))

	rand.Shuffle(len(words), func(i, j int) {
		words[i], words[j] = words[j], words[i]
	})
}

func GetQuote(path string) string {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	words := make([]string, 0)

	for scanner.Scan() {
		words = append(words, scanner.Text())
	}

	words = Filter(words, func(w string) bool {
		return !strings.Contains(w, "'") && 5 <= len(w) && len(w) <= 6
	})

	for i := 0; i < len(words); i++ {
		words[i] = strings.ToLower(words[i])
	}

	shuffle(words)

	return strings.Join(words, " ")
}
