package app

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"

	"DelayedNotifier/internal/repository"
	"DelayedNotifier/internal/service"
)

type App struct {
	Host     string
	Handler  http.Handler
	RabbitMQ *repository.RabbitMQ
	DB       *repository.Storage
	TgBot    *service.TelegramBot
	Redis    *redis.Client
}

func (a *App) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	msgChan := make(chan []byte, 100)

	retryQueue := make(chan service.RetryMessage, 100) // сюда отправляются сообщения уже на отправку в тг

	server := http.Server{
		Addr:    ":" + a.Host,
		Handler: a.Handler,
	}

	zlog.Logger.Info().Msgf("Server start with addr: %s", a.Host)

	wg.Add(1)
	go func() {

		defer func() {
			zlog.Logger.Info().Msg("завершение сервера")
			wg.Done()
		}()
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			zlog.Logger.Err(err)
			cancel()
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer func() {
			zlog.Logger.Info().Msg("завершение консумера")
			wg.Done()
		}()
		strConsumer := retry.Strategy{
			Attempts: 5,
			Delay:    5 * time.Second,
			Backoff:  2,
		}

		errChan := make(chan error, 1)
		go func() {
			errChan <- a.RabbitMQ.Consumer.ConsumeWithRetry(msgChan, strConsumer)
		}()

		select {
		case err := <-errChan:
			if err != nil {
				zlog.Logger.Error().Msgf("ConsumeWithRetry error: %v", err)
				cancel()
			}
		case <-ctx.Done():
			zlog.Logger.Info().Msg("Context cancelled, stopping consumer")
		}

		close(msgChan)
	}()

	wg.Add(1)
	go func() {
		defer func() {
			zlog.Logger.Info().Msg("завершение отправки сообещений")
			wg.Done()
		}()
		service.SendMessage(ctx, retryQueue, msgChan, a.DB)
	}()

	wg.Add(1)
	go func() {
		defer func() {
			zlog.Logger.Info().Msg("завершение прослушки бота")
			wg.Done()
		}()
		if a.TgBot != nil {
			a.TgBot.ListenUpdated(ctx)

		}
	}()

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer func() {
				zlog.Logger.Info().Msg("завершение воркера")
				wg.Done()
			}()
			service.RetryWorker(ctx, retryQueue, a.DB, a.TgBot)
		}()
	}

	<-ctx.Done()

	zlog.Logger.Info().Msg("Shutdown app..")

	err := server.Shutdown(ctx)
	if err != nil {
		zlog.Logger.Error().Msgf("error in close server: %v", err)
	}

	wg.Wait()

	err = a.DB.Close()
	if err != nil {
		zlog.Logger.Error().Msgf("error in close DB: %v", err)
	}

	err = a.Redis.Close()
	if err != nil {
		zlog.Logger.Error().Msgf("error in close Redis: %v", err)

	}

	return nil

}
