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
	zerolog.TimeFieldFormat = time.RFC3339Nano
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.StampMilli}
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

	log.Info().Msg("start cache warm up")
	repo.СacheWarmUp()
	// Если не ждать процесс заполнения кеша
	// Это может вытеснить более свежие заказы добавленные
	// ресивером более старыми из базы данных
	log.Info().Msg("done cache warm up")

	rec, err := receiver.New(repo, cfg, log)
	if err != nil {
		log.Fatal().Err(err).Msg("fail new receiver")
	}
	defer rec.Close()
    
	err = rec.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("fail run receiver")
	}

	e := endpoint.New(repo, log)
	go e.Run()

	log.Info().Msg("starting http service")

	signals := make(chan os.Signal, 2)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT)

	for {
		sign := <-signals
		ctxCancel()
		log.Info().Str("signal", sign.String()).Msg("stoping service")
		break
	}
}
