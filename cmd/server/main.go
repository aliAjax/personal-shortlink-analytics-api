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

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	accesshandler "github.com/example/shortlink-api/internal/access/handler"
	accessrepository "github.com/example/shortlink-api/internal/access/repository"
	accessservice "github.com/example/shortlink-api/internal/access/service"
	redirecthandler "github.com/example/shortlink-api/internal/redirect/handler"
	redirectrepository "github.com/example/shortlink-api/internal/redirect/repository"
	redirectservice "github.com/example/shortlink-api/internal/redirect/service"
	shortlinkhandler "github.com/example/shortlink-api/internal/shortlink/handler"
	shortlinkmodel "github.com/example/shortlink-api/internal/shortlink/model"
	shortlinkrepository "github.com/example/shortlink-api/internal/shortlink/repository"
	shortlinkservice "github.com/example/shortlink-api/internal/shortlink/service"
	userhandler "github.com/example/shortlink-api/internal/user/handler"
	usermodel "github.com/example/shortlink-api/internal/user/model"
	userrepository "github.com/example/shortlink-api/internal/user/repository"
	userservice "github.com/example/shortlink-api/internal/user/service"
	"github.com/example/shortlink-api/pkg/config"
	"github.com/example/shortlink-api/pkg/database"
	"github.com/example/shortlink-api/pkg/middleware"
)

func main() {
	if err := run(); err != nil {
		slog.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := config.Load()
	db, err := database.Open(cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := database.WaitReady(ctx, db, 90*time.Second); err != nil {
		return err
	}
	slog.Info("database is ready")

	if err := database.Migrate(ctx, db, cfg.MigrationsDir); err != nil {
		return err
	}
	slog.Info("database migrations applied")

	userRepo := userrepository.NewRepository(db)
	userService := userservice.NewService(userRepo)
	shortlinkRepo := shortlinkrepository.NewRepository(db)
	shortlinkService := shortlinkservice.NewService(shortlinkRepo)
	accessRepo := accessrepository.NewRepository(db)
	accessService := accessservice.NewService(accessRepo)
	redirectRepo := redirectrepository.NewRepository(db)
	redirectService := redirectservice.NewService(redirectRepo, accessService)

	if err := seedDemo(ctx, userService, userRepo, shortlinkService); err != nil {
		return err
	}

	userHandler := userhandler.NewHandler(userService, cfg.JWTSecret, cfg.JWTExpiresHours)
	shortlinkHandler := shortlinkhandler.NewHandler(shortlinkService)
	accessHandler := accesshandler.NewHandler(accessService, shortlinkService)
	redirectHandler := redirecthandler.NewHandler(redirectService)

	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(chimiddleware.Logger)
	router.Use(chimiddleware.Recoverer)

	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	userHandler.RegisterRoutes(router)
	redirectHandler.RegisterRoutes(router)

	router.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.Auth(cfg.JWTSecret))
		shortlinkHandler.RegisterRoutes(r)
		accessHandler.RegisterRoutes(r)
	})

	server := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	slog.Info("server listening", "port", cfg.ServerPort)
	return server.ListenAndServe()
}

func seedDemo(
	ctx context.Context,
	userService *userservice.Service,
	userRepo userrepository.Repository,
	shortlinkService *shortlinkservice.Service,
) error {
	user, err := userService.Register(ctx, usermodel.RegisterRequest{
		Username: "demo",
		Password: "demo123456",
	})
	if err != nil && !errors.Is(err, userrepository.ErrDuplicateUsername) {
		return err
	}
	if errors.Is(err, userrepository.ErrDuplicateUsername) {
		user, err = userRepo.FindByUsername(ctx, "demo")
		if err != nil {
			return err
		}
	}

	links, err := shortlinkService.List(ctx, user.ID)
	if err != nil {
		return err
	}
	if len(links) > 0 {
		return nil
	}

	if _, err := shortlinkService.Create(ctx, user.ID, shortlinkmodel.CreateRequest{
		OriginalURL: "https://example.com/demo",
		CustomAlias: "demo-home",
	}); err != nil {
		return err
	}
	expired := time.Now().Add(-time.Hour)
	if _, err := shortlinkService.Create(ctx, user.ID, shortlinkmodel.CreateRequest{
		OriginalURL: "https://example.com/expired",
		CustomAlias: "demo-expired",
		ExpiresAt:   &expired,
	}); err != nil {
		return err
	}
	return nil
}
