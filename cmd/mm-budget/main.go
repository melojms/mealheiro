// Command mm-budget runs the household budget web app.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/melojms/mm-budget/internal/api"
	"github.com/melojms/mm-budget/internal/backup"
	"github.com/melojms/mm-budget/internal/clock"
	"github.com/melojms/mm-budget/internal/recurring"
	"github.com/melojms/mm-budget/internal/store"
	"github.com/melojms/mm-budget/web"
)

const backupsToKeep = 30

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(healthcheck())
	}
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	level := slog.LevelInfo
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	slog.SetDefault(log)

	addr := env("ADDR", ":8080")
	dataDir := env("DATA_DIR", "./data")
	tz := env("TZ", "Europe/Lisbon")

	loc, err := time.LoadLocation(tz)
	if err != nil {
		return err
	}
	clk := clock.System(loc)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := store.Open(ctx, filepath.Join(dataDir, "mm-budget.db"))
	if err != nil {
		return err
	}
	defer db.Close()

	backupDir := filepath.Join(dataDir, "backups")
	go runDaily(ctx, log, "recurring", func() error {
		n, err := recurring.Generate(ctx, db, clk)
		if n > 0 {
			log.Info("recurring entries generated", "count", n)
		}
		return err
	})
	go runDaily(ctx, log, "backup", func() error {
		return backup.Nightly(ctx, db, backupDir, clock.Today(clk), backupsToKeep)
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           api.New(db, clk, log, backupDir, web.Dist()).Routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		log.Info("listening", "addr", addr, "data_dir", dataDir, "tz", tz, "version", api.Version)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		log.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
	return nil
}

// runDaily runs fn now and then every hour; fn implementations are idempotent
// per day/month, so the hourly tick just guarantees a run soon after midnight.
func runDaily(ctx context.Context, log *slog.Logger, name string, fn func() error) {
	tick := time.NewTicker(time.Hour)
	defer tick.Stop()
	for {
		if err := fn(); err != nil {
			log.Error("job failed", "job", name, "err", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}
	}
}

// healthcheck probes the local /healthz endpoint (used by Docker; the image has no curl).
func healthcheck() int {
	port := strings.TrimPrefix(env("ADDR", ":8080"), ":")
	if i := strings.LastIndex(port, ":"); i >= 0 {
		port = port[i+1:]
	}
	client := http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://127.0.0.1:" + port + "/healthz")
	if err != nil {
		return 1
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
