package main

import (
    "context"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/seu-user/url-shortener/internal/config"
    "github.com/seu-user/url-shortener/internal/handler"
    "github.com/seu-user/url-shortener/internal/repository"
    "github.com/seu-user/url-shortener/internal/service"
)

func main() {
    cfg := config.Load()

    pgRepo := repository.NewPostgres(cfg.DatabaseURL)
    redisRepo := repository.NewRedis(cfg.RedisURL)
    svc := service.NewURLService(pgRepo, redisRepo)
    h := handler.NewURLHandler(svc)

    mux := http.NewServeMux()
    mux.HandleFunc("POST /shorten", h.Shorten)
    mux.HandleFunc("GET /{code}", h.Redirect)
    mux.HandleFunc("GET /health", h.Health)

    srv := &http.Server{
        Addr:    ":" + cfg.Port,
        Handler: mux,
    }

    go func() {
        slog.Info("server starting", "port", cfg.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            slog.Error("server error", "err", err)
            os.Exit(1)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    srv.Shutdown(ctx)
}