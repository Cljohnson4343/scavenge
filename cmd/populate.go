package cmd

import (
	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/populate"
	"github.com/spf13/cobra"
)

var populateCmd = &cobra.Command{
	Use:       "populate [testing | production | development]",
	Short:     "populate the database for the given environment",
	Args:      cobra.ExactValidArgs(1),
	ValidArgs: []string{"production", "testing", "development"},
	Run: func(cmd *cobra.Command, args []string) {
		deployEnv := args[0]

		database := db.InitDB(deployEnv)
		defer db.Shutdown(database)

		populate.Populate(true)

	},
}

func init() {
	rootCmd.AddCommand(populateCmd)
	populateCmd.Flags()

}
