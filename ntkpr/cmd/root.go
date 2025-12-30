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

var globalCfg *config.Config
var globalDB *db.DB
var globalApp *app.App
var globalModel *ui.Model

var rootCmd = &cobra.Command{
	Use:   "ntkpr",
	Short: "ntkpr",
	Long:  "ntkpr",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Load app config
		cfg := config.LoadOrCreateConfig()
		globalCfg = &cfg

		// Initialize database
		var err error
		globalDB, err = db.NewDB(cfg.DataFilePath + "/notes.db")
		if err != nil {
			log.Fatal("Failed to connect to database:", err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Get state
		s, err := state.LoadState(globalCfg.StateFilePath)
		if err != nil {
			log.Fatal("Failed to load state:", err)
		}

		// Initialize application
		globalApp = app.NewApp(globalDB)

		// Initialize UI model
		model := ui.NewModel(globalApp, globalCfg, s)
		globalModel = &model

		// Run Bubble Tea program
		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}
	},
}

func Execute() {
	defer func() {
		if globalDB != nil {
			globalDB.Close()
		}
	}()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Oops. An error while executing Zero '%s'\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(ExportNoteCmd)
	rootCmd.AddCommand(LaunchGUICmd)
	rootCmd.AddCommand(DataBackupCmd)
}
