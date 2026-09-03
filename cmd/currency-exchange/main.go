package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"github.com/Endcru/PlataTestAssignment/docs"

	"github.com/Endcru/PlataTestAssignment/internal/config"
	"github.com/Endcru/PlataTestAssignment/internal/storage/postgres"
	"github.com/Endcru/PlataTestAssignment/internal/http-server/middleware/logger"
	quotationService "github.com/Endcru/PlataTestAssignment/internal/service/quotation"
	quotationAPIService "github.com/Endcru/PlataTestAssignment/internal/service/quotationAPI"
	quotationUpdateSheduler "github.com/Endcru/PlataTestAssignment/internal/sheduler/quotationUpdateSheduler"
	quotation "github.com/Endcru/PlataTestAssignment/internal/http-server/handlers/url/quotation"
)

const (
	envLocal = "local"
	envDev = "dev"
	envProd = "prod"
)

// @title Currency Exchange API
// @version 1.0
// @description Asynchronous currency quotation service
// @host localhost:8090
// @BasePath /
func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	log = log.With(slog.String("env", cfg.Env))
	log.Info("Initialize server", slog.String("address", cfg.HTTPServer.Address))

	// Подключаем storage
	storage, err := postgres.NewPostgresStorage(cfg)
	if err != nil {
		log.Error("Failed to create storage", "error", err)
		os.Exit(1)
	}
	log.Info("Storage created successfully")
	defer storage.Close()

	// Добавляем базовые запросы в storage
	err = postgres.AddBasePostgresStatements(storage)
	if err != nil {
		log.Error("Failed to add base statements", "error", err)
		os.Exit(1)
	}
	log.Info("Base statements added successfully")

	// Создаем сервисы
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Error("API_KEY is not set")
		os.Exit(1)
	}
	quotationAPIService := quotationAPIService.NewQuotationAPIService("currencybeacon", apiKey, log)
	quotationService := quotationService.NewQuotationService(storage, log, quotationAPIService)

	err = quotationService.CreateStartQuotations()
	if err != nil {
		log.Error("Failed to create start quotations", "error", err)
		os.Exit(1)
	}
	log.Info("Start quotations created successfully")

	// Подключаем роутеры и middleware

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer) 
	router.Use(middleware.URLFormat)

	router.Use(logger.New(log))

	router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.InstanceName(docs.SwaggerInfo.InstanceName()),
	))

	router.Route("/quotation", func(r chi.Router) {
		r.Get("/{name}", quotation.GetQuotation(log, quotationService))
		r.Post("/{name}/update", quotation.QuotationUpdate(log, quotationService))
		r.Get("/request/{update_id}", quotation.GetQuotationFromRequestUpdateID(log, quotationService))
		r.Get("/{name}/updates", quotation.GetQuotationUpdates(log, quotationService))
	})

	log.Info("Server starting server")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr: cfg.HTTPServer.Address,
		Handler: router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	log.Info("Server started")

	updateTime := cfg.QuotationUpdateScheduler.Duration


	quotationUpdateSheduler := quotationUpdateSheduler.NewQuotationUpdateSheduler(quotationService, log, updateTime)
	quotationUpdateSheduler.Start()
	defer quotationUpdateSheduler.Stop()

	<-done
	log.Info("Server stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Failed to shutdown server", "error", err)
		os.Exit(1)
	}

	defer storage.Close()

	log.Info("Server exited")

	os.Exit(0)
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, nil))
	case envDev:
		log = slog.New(slog.NewTextHandler(os.Stdout, nil))
	case envProd:
		log = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}
	return log
}