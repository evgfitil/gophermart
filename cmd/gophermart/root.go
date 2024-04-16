package main

import (
	"github.com/caarlos0/env/v10"
	"github.com/evgfitil/gophermart.git/internal/database"
	"github.com/evgfitil/gophermart.git/internal/logger"
	"github.com/evgfitil/gophermart.git/internal/router"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	defaultRunAddress = "localhost:8080"
)

var (
	cfg     *Config
	rootCmd = &cobra.Command{
		Use:   "server",
		Short: "Gophermart Loyalty System",
		Long: `Gophermart Loyalty System is a comprehensive server-side application designed to manage a rewards-based loyalty program. 
                This system allows registered users to submit order numbers, tracks these submissions, 
                and interfaces with an external accrual system to calculate loyalty points based on user purchases.`,
		Run: runServer,
	}
)

func runServer(cmd *cobra.Command, args []string) {
	logger.InitLogger(cfg.LogLevel)
	defer logger.Sugar.Sync()

	if err := env.Parse(cfg); err != nil {
		logger.Sugar.Fatalf("error parsing config: %v", err)
	}

	db, err := database.NewDBStorage(cfg.DatabaseURI)
	if err != nil {
		logger.Sugar.Fatalf("error connecting to database: %v", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Sugar.Infoln("starting server")
		err := http.ListenAndServe(cfg.RunAddress, router.ApiRouter(*db))
		if err != nil {
			logger.Sugar.Fatalf("error starting server: %v", err)
		}
	}()
	<-quit
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cfg = NewConfig()
	rootCmd.Flags().StringVarP(&cfg.RunAddress, "address", "a", defaultRunAddress, "run address for the server in the format host:port")
	rootCmd.Flags().StringVarP(&cfg.DatabaseURI, "database-uri", "d", "", "database connection string")
	rootCmd.Flags().StringVarP(&cfg.AccrualSystemAddress, "accrual-system-address", "r", "", "accrual system address")
}
