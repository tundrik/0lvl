package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"0lvl/config"
	"0lvl/internal/endpoint"
	"0lvl/internal/receiver"
	"0lvl/internal/repository"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/rs/zerolog"
)

func createLogger() zerolog.Logger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	return zerolog.New(output).With().Timestamp().Logger()
}

func main() {
	ctx, ctxCancel := context.WithCancel(context.Background())
	log := createLogger()

	var cfg config.Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatal().Err(err).Msg("fail read config")
	}

	repo, err := repository.New(ctx, cfg, log)
	if err != nil {
		log.Fatal().Err(err).Msg("fail new repository")
	}

	conn, err := receiver.New("3", repo, cfg, log)
	if err != nil {
		log.Fatal().Err(err).Msg("fail new receiver")
	}
	defer conn.Close()

	e := endpoint.New(repo, log)
	go e.Run()

	log.Info().Msg("starting service")

	signals := make(chan os.Signal, 2)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

	for {
		sign := <-signals
		ctxCancel()
		log.Info().Str("signal", sign.String()).Msg("stoping service")
		break
	}
}
