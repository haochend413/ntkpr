package cmd

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/haochend413/ntkpr/internal/app"
	"github.com/haochend413/ntkpr/internal/db"
	"github.com/haochend413/ntkpr/internal/ui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ntkpr",
	Short: "ntkpr",
	Long:  "ntkpr",
	Run: func(cmd *cobra.Command, args []string) {

		// Initialize database
		dbConn, err := db.NewDB("notes.db")
		if err != nil {
			log.Fatal("Failed to connect to database:", err)
		}
		defer dbConn.Close()

		// Initialize application
		application := app.NewApp(dbConn)

		// Initialize UI model
		model := ui.NewModel(application)

		// Run Bubble Tea program
		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Oops. An error while executing Zero '%s'\n", err)
		os.Exit(1)
	}
}
