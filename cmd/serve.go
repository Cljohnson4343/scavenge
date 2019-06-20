package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cljohnson4343/scavenge/config"
	"github.com/cljohnson4343/scavenge/db"
	"github.com/cljohnson4343/scavenge/response"
	"github.com/cljohnson4343/scavenge/routes"
	"github.com/go-chi/chi"
	"github.com/spf13/cobra"
)

var devModeFlag *bool
var portFlag *int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Serve the scavenge api",
	Run: func(cmd *cobra.Command, args []string) {
		deployEnv := "production"
		if *devModeFlag {
			response.SetDevMode(true)
			deployEnv = "development"
		}

		database := db.InitDB(deployEnv)
		defer db.Shutdown(database)

		env := config.CreateEnv(database)
		router := routes.Routes(env)

		walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
			log.Printf("%s %s\n", method, route) // walk and print out all routes
			return nil
		}
		if err := chi.Walk(router, walkFunc); err != nil {
			log.Panicf("Logging err: %s\n", err.Error()) // panic if there is an error
		}

		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *portFlag), router))

	},
}

func init() {
	devModeFlag = serveCmd.PersistentFlags().Bool("dev-mode", false, "set the server to dev mode")
	portFlag = serveCmd.Flags().Int("port", 4343, "port to use")

	rootCmd.AddCommand(serveCmd)
}
