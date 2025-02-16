package main

import (
	"net/http"
	"time"

	sqlitedb "go.altair.com/todolist/pkg/db"
	"go.altair.com/todolist/pkg/todolist"
	"go.altair.com/todolist/pkg/todolist/store"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Runs the Todolist server",
	Long:  ``,
	RunE:  doServe,
}

var (
	bindAddress string
)

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVarP(&bindAddress, "bind", "b", "0.0.0.0:8080", "set the bind address for the server")
}

func newRouter() *chi.Mux {
	router := chi.NewRouter()
	router.Use(chimw.Recoverer)
	router.Use(chimw.Timeout(60 * time.Second))
	return router
}

func doServe(cmd *cobra.Command, args []string) error {
	log.Info().Msg(description + " starting")

	tododb, err := sqlitedb.CreateDb()
	if err != nil {
		return err
	}

	todostore := store.NewSqlStore(tododb)
	todoService := todolist.NewItemsService(todostore)

	handler := &todolist.ItemsHandlers{
		ItemsService: todoService,
	}

	router := newRouter()
	handler.ConfigureRoutes(router)

	log.Info().Str("bindAddress", bindAddress).Msg("Listening for HTTP requests")
	return http.ListenAndServe(bindAddress, router)
}
