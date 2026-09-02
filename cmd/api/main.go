package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Kaveesha23dil/go-task-platform/internal/handler"
	"github.com/Kaveesha23dil/go-task-platform/internal/queue"
	"github.com/Kaveesha23dil/go-task-platform/internal/repository"
	"github.com/Kaveesha23dil/go-task-platform/internal/service"
	"github.com/Kaveesha23dil/go-task-platform/internal/worker"
	"github.com/gin-gonic/gin"
)

const (
	workerCount      = 3
	queueSize        = 100
	processingTime   = 3 * time.Second
	shutdownDeadline = 10 * time.Second
)

func main() {
	logger := slog.Default()
	router := gin.Default()
	taskRepository := repository.NewMemoryTaskRepository()
	taskQueue := queue.NewLocalQueue(queueSize)
	taskService := service.NewTaskService(taskRepository, taskQueue)
	taskHandler := handler.NewTaskHandler(taskService)
	handler.RegisterRoutes(router, taskHandler)

	pool := worker.NewPool(workerCount, taskQueue, taskService, worker.SimulatedProcessor(processingTime), logger)
	appContext, cancelWorkers := context.WithCancel(context.Background())
	defer cancelWorkers()
	if err := pool.Start(appContext); err != nil {
		logger.Error("could not start worker pool", "error", err)
		return
	}

	server := &http.Server{Addr: ":8080", Handler: router, ReadHeaderTimeout: 5 * time.Second}
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- server.ListenAndServe() }()

	signalContext, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()
	select {
	case <-signalContext.Done():
		logger.Info("shutdown signal received")
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server stopped unexpectedly", "error", err)
		}
	}

	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), shutdownDeadline)
	defer cancelShutdown()
	if err := server.Shutdown(shutdownContext); err != nil {
		logger.Error("HTTP server shutdown failed", "error", err)
	}
	if err := pool.Shutdown(shutdownContext); err != nil {
		logger.Error("worker pool shutdown failed", "error", err)
	}
}
