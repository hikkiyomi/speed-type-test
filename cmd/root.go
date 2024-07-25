package cmd

import (
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	stt "github.com/hikkiyomi/speed-type-test/internal"
)

const (
	DEBUG_FILE   = "debug.log"
	DEBUG_PREFIX = "DEBUG"
)

var rootCmd = &cobra.Command{
	Use:   "stt",
	Short: "Type stt to start the application.",
	Run: func(cmd *cobra.Command, args []string) {
		f, err := tea.LogToFile(DEBUG_FILE, DEBUG_PREFIX)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		timeoutInSeconds, err := cmd.Flags().GetInt("timeout")
		if err != nil {
			log.Fatal("Couldn't get the 'timeout' flag.")
		}

		wrapWords, err := cmd.Flags().GetInt("wrap")
		if err != nil {
			log.Fatal("Couldn't get the 'wrap' flag.")
		}

		input, err := cmd.Flags().GetString("input")
		if err != nil {
			log.Fatal("Couldn't get the 'input' flag")
		}

		// A quote is just a space-separated set of random intelligible words.
		quote := stt.GetQuote(input)

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
	rootCmd.Flags().IntP("timeout", "t", 30, "specifies the timeout for timer. Put 0 for infinite amount of time.")
	rootCmd.Flags().IntP("wrap", "w", 10, "specifies the number of words in one line.")
}
