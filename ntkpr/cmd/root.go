package cmd

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/haochend413/ntkpr/config"
	"github.com/haochend413/ntkpr/internal/app"
	"github.com/haochend413/ntkpr/internal/db"
	"github.com/haochend413/ntkpr/internal/ui"
	"github.com/haochend413/ntkpr/state"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ntkpr",
	Short: "ntkpr",
	Long:  "ntkpr",
	Run: func(cmd *cobra.Command, args []string) {

		// load app config
		cfg := config.LoadOrCreateConfig()

		// Initialize database
		fmt.Printf(cfg.DataFilePath)
		fmt.Printf(cfg.DataFilePath + "/notes.db")
		dbConn, err := db.NewDB(cfg.DataFilePath + "/notes.db")
		if err != nil {
			log.Fatal("Failed to connect to database:", err)
		}
		defer dbConn.Close()

		//get state
		s, err := state.LoadState(cfg.StateFilePath)
		if err != nil {
			log.Fatal("Failed to load state:", err)
		}

		// Initialize application
		application := app.NewApp(dbConn)

		// Initialize UI model
		model := ui.NewModel(application, &cfg, s)

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
