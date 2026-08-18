package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/carlosSimplicio/judcalc/backend/internal/httpapi"
	sqlitestorage "github.com/carlosSimplicio/judcalc/backend/internal/storage/sqlite"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("API encerrada", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	databasePath := environmentOrDefault("DATABASE_PATH", "data/app.db")
	httpAddress := environmentOrDefault("HTTP_ADDR", ":8080")

	startupContext, cancelStartup := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelStartup()
	database, err := sqlitestorage.OpenDatabase(startupContext, databasePath)
	if err != nil {
		return err
	}
	defer database.Close()

	areaRepository := sqlitestorage.NewAreaRepository(database)
	serviceRepository := sqlitestorage.NewServiceRepository(database)
	fixedCostsRepository := sqlitestorage.NewFixedCostsRepository(database)
	server := &http.Server{
		Addr:              httpAddress,
		Handler:           httpapi.NewRouter(areaRepository, serviceRepository, fixedCostsRepository, logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	logger.Info("API iniciada", "address", httpAddress, "database", databasePath)
	return server.ListenAndServe()
}

func environmentOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
