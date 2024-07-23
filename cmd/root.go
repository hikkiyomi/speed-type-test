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

		quote := "somebody once told me hello everybody this is my program for speed testing"

		timeoutInSeconds, err := cmd.Flags().GetInt("timeout")
		if err != nil {
			log.Fatal("Couldn't get the timeout flag.")
		}

		p := tea.NewProgram(stt.NewModel(quote, timeoutInSeconds), tea.WithAltScreen())

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
	rootCmd.Flags().IntP("timeout", "t", 30, "Specifies the timeout for timer. Put 0 for infinite amount of time.")
}
