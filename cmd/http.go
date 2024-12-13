/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/okiww/billing-system-okky/configs"
	"github.com/okiww/billing-system-okky/pkg/db"
	"github.com/okiww/billing-system-okky/pkg/logger"
	"github.com/okiww/billing-system-okky/port/rest"
	"github.com/spf13/cobra"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// httpCmd represents the http command
var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		ServeHttp()
	},
}

func init() {
	rootCmd.AddCommand(httpCmd)
}

func ServeHttp() {
	router := mux.NewRouter()
	rest.RegisterRoutes(router)
	cfg := configs.InitConfig()
	db.InitDB(cfg.DB.Source)

	// Create the HTTP server
	server := &http.Server{
		Addr:    cfg.Http.Addr,
		Handler: router,
	}

	// Run the server in a separate goroutine
	go func() {
		log.Println("Server running on port", cfg.Http.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Set up channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Block until a signal is received
	<-stop
	log.Println("Shutting down server...")

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	defer func() {
		// extra handling here
		err := db.CloseDB()
		if err != nil {
			logger.Fatalf("failed close db %s", err.Error())
		}
		serverStopCtx()
		<-serverCtx.Done()
	}()

	// Attempt a graceful shutdown
	if err := server.Shutdown(serverCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
