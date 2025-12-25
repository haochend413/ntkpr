package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var LaunchGUICmd = &cobra.Command{
	Use:   "gui",
	Short: "Launch Gui. ",
	Long:  "Launch the Gui for lastest exported notes.",
	Run: func(cmd *cobra.Command, args []string) {
		//fetch from globaldb
		cmdd := exec.Command("pnpm", "dev")
		guiDir, err := filepath.Abs("../gui")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error resolving gui path: %v\n", err)
			return
		}
		// set working directory
		cmdd.Dir = guiDir
		cmdd.Stdout = os.Stdout
		cmdd.Stderr = os.Stderr
		fmt.Println("Starting GUI at:", guiDir)
		if err := cmdd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running pnpm dev: %v\n", err)
			return
		}
	},
}
