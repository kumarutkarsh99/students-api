package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kumarutkarsh99/students-api/internal/config"
	"github.com/kumarutkarsh99/students-api/internal/http/handlers/student"
	"github.com/kumarutkarsh99/students-api/internal/storage/sqlite"
)

func main() {
	// load config
	cfg := config.MustLoad()

	//database setup
	storage, err := sqlite.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := storage.Close(); err != nil {
			slog.Error("failed to close storage", slog.String("error", err.Error()))
		}
	}()

	slog.Info("storage initialised", slog.String("env", cfg.Env))

	//setup router
	router := http.NewServeMux()
	router.HandleFunc("POST /api/students", student.New(storage))
	router.HandleFunc("GET /api/students/{id}", student.GetStudentById(storage))
	router.HandleFunc("GET /api/students", student.GetList(storage))
	router.HandleFunc("PUT /api/students/{id}", student.UpdateStudentById(storage))
	router.HandleFunc("DELETE /api/students/{id}", student.DeleteStudentById(storage))

	//setup server
	server := http.Server{
		Addr:              cfg.Addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	slog.Info("server started", slog.String("address", cfg.Addr))

	done := make(chan os.Signal, 1)
	serverErrors := make(chan error, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(done)

	go func() {
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	select {
	case <-done:
	case err := <-serverErrors:
		slog.Error("failed to start server", slog.String("error", err.Error()))
		return
	}

	slog.Info("shutting down the server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("failed to shutdown server", slog.String("error", err.Error()))
	}

	slog.Info("server shutdown successfully")
}
