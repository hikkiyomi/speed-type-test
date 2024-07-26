package cmd

import (
	"log"
	"os"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	stt "github.com/hikkiyomi/speed-type-test/internal"
)

const (
	DEBUG_FILE   = "debug.log"
	DEBUG_PREFIX = "DEBUG"
)

func GetFlag[T int | string](cmd *cobra.Command, name string) T {
	value := cmd.Flags().Lookup(name).Value

	var result any
	var err error

	switch value.Type() {
	case "int":
		result, err = strconv.Atoi(value.String())
	case "string":
		result = value.String()
	}

	if err != nil {
		log.Fatal(err)
	}

	actualValue, ok := result.(T)

	if !ok {
		log.Fatalf("Couldn't get '%s'", name)
	}

	return actualValue
}

var rootCmd = &cobra.Command{
	Use:   "stt",
	Short: "Type stt to start the application.",
	Run: func(cmd *cobra.Command, args []string) {
		timeoutInSeconds := GetFlag[int](cmd, "timeout")
		wrapWords := GetFlag[int](cmd, "wrap")
		input := GetFlag[string](cmd, "input")
		minLength := GetFlag[int](cmd, "minlen")
		maxLength := GetFlag[int](cmd, "maxlen")

		// A quote is just a space-separated set of random intelligible words.
		quote := stt.GetQuote(input, minLength, maxLength)

		p := tea.NewProgram(stt.NewModel(quote, timeoutInSeconds, wrapWords), tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("input", "i", "/usr/share/dict/american-english", "path for word collection to train on.")
	rootCmd.Flags().IntP("timeout", "t", 30, "specifies the timeout (in seconds) for timer. Put 0 for infinite amount of time.")
	rootCmd.Flags().IntP("wrap", "w", 10, "specifies the number of words in one line.")
	rootCmd.Flags().Int("minlen", 5, "specifies the minimum length of words. Put 0 for no limit.")
	rootCmd.Flags().Int("maxlen", 6, "specifies the maximum length of words. Put 0 for no limit.")
}
