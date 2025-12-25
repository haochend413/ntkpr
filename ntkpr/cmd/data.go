package cmd

import (
	"github.com/spf13/cobra"
)

var ExportNoteCmd = &cobra.Command{
	Use:   "export",
	Short: "export",
	Long:  "export",
	Run: func(cmd *cobra.Command, args []string) {
		//fetch from globaldb
		globalDB.ExportNoteToJSON(globalCfg.DataFilePath + "/notes.json")
	},
}
