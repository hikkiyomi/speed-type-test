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

		quote := "somebody once told me"

		if _, err := tea.NewProgram(stt.NewModel(quote)).Run(); err != nil {
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
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
