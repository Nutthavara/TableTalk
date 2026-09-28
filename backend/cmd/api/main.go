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
	_ "time/tzdata" // restaurant timezones must resolve even where the OS has no tz database

	"tabletalk/backend/internal/auth"
	"tabletalk/backend/internal/config"
	"tabletalk/backend/internal/database"
	"tabletalk/backend/internal/handler"
	"tabletalk/backend/internal/middleware"
	"tabletalk/backend/internal/repository"
	"tabletalk/backend/internal/service"
	"tabletalk/backend/internal/storage"
)

func main() {
	if err := run(); err != nil {
		slog.Error("api stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	verifier := auth.NewKeycloakVerifier(ctx, cfg.OIDCIssuerURL, cfg.OIDCAudience)
	store := repository.NewStore(pool)
	files, err := storage.NewLocal(cfg.UploadDir, cfg.PublicBaseURL+"/uploads")
	if err != nil {
		return err
	}

	reviews := service.NewReviewService(store)
	router := handler.NewRouter(handler.Deps{
		DB:           pool,
		Auth:         middleware.NewAuthenticator(verifier, store),
		Restaurants:  handler.NewRestaurantHandler(service.NewRestaurantService(store, files, time.Now), reviews),
		Reservations: handler.NewReservationHandler(service.NewReservationService(store, time.Now)),
		Reviews:      handler.NewReviewHandler(reviews),
		Uploads:      files.Handler(),
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		// Generous enough for multi-image uploads; bounds slow clients holding connections open.
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 2 * time.Minute,
		IdleTimeout:  2 * time.Minute,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("api listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}
