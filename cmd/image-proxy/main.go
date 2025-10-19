package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shuimo03/image-proxy/internal/config"
	"github.com/Shuimo03/image-proxy/internal/logging"
	"github.com/Shuimo03/image-proxy/internal/server"
	"github.com/Shuimo03/image-proxy/internal/transport"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "./configs/example.toml", "path to TOML configuration file")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := logging.New()

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("failed to load config", "path", configPath, "err", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		logger.Error("invalid configuration", "err", err)
		os.Exit(1)
	}

	transport, err := transport.New(cfg)
	if err != nil {
		logger.Error("failed to configure transport", "err", err)
		os.Exit(1)
	}

	srv, err := server.New(cfg, transport, logger)
	if err != nil {
		logger.Error("failed to init server", "err", err)
		os.Exit(1)
	}

	if err := srv.Start(ctx); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server encountered error", "err", err)
			os.Exit(1)
		}
		logger.Info("server exited cleanly")
	} else {
		logger.Info("server stopped")
	}
}
