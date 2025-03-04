package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/app/clients"
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/app/config"
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/app/repository/pgrepo"
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/app/services"
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/app/transport/httpserver"
	"github.com/yuriyfomin17/largest-picture-nasa-api/internal/pkg"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

func run() error {
	cfg := config.Read()
	mq, err := pkg.ConnectRabbitMQ(cfg.RabbitMQURL)
	if err != nil {
		return fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	pgDB, err := pkg.Dial(cfg.DSN)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	if pgDB != nil {
		log.Println("Running Postgres migrations")
		if err := runPgMigrations(cfg.DSN, cfg.MigrationsPath); err != nil {
			return fmt.Errorf("runPgMigrations failed: %w", err)
		}
	}

	pictureRepo := pgrepo.NewPictureRepo(pgDB)

	nasaApiClient := clients.NewNasaApiClient(cfg.APIKey, cfg.APIUrl)
	largestPictureService := services.NewLargestPictureService(
		mq,
		&pictureRepo,
		nasaApiClient,
	)
	largestPictureService.StartListeningSolCommands()

	largestPictureServer := httpserver.NewHttpServer(largestPictureService)

	// create http router
	router := mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("NASA Largest Picture API 1.0V"))
	}).Methods("GET")

	router.HandleFunc("/mars/pictures/largest/command", largestPictureServer.PostCommandHandler).Methods("POST")
	router.HandleFunc("/mars/pictures/largest/command/{sol}", largestPictureServer.GetLargestPictureHandler).Methods("GET")

	srv := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: router,
	}
	// listen to OS signals and gracefully shutdown HTTP server
	stopped := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-sigint
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("HTTP Server Shutdown Error: %v", err)
		}
		close(stopped)
	}()

	log.Printf("Starting HTTP server on %s", cfg.HTTPAddr)

	// start HTTP server
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP server ListenAndServe Error: %v", err)
	}

	<-stopped

	log.Printf("Have a nice day!")
	return nil
}

// runPgMigrations runs Postgres migrations
func runPgMigrations(dsn, path string) error {
	if path == "" {
		return errors.New("no migrations path provided")
	}
	if dsn == "" {
		return errors.New("no DSN provided")
	}

	m, err := migrate.New(
		path,
		dsn,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}
