package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/zlog"

	"DelayedNotifier/internal/api"
	"DelayedNotifier/internal/api/handlers"
	"DelayedNotifier/internal/app"
	"DelayedNotifier/internal/config"
	"DelayedNotifier/internal/repository"
	"DelayedNotifier/internal/service"
)

func main() {

	zlog.Init()

	c, err := config.NewConfig()
	if err != nil {
		zlog.Logger.Fatal().Msgf("error load config: %v", err)
		return
	}
	// Добавил возможность не прикручивать бота, чтобы можно было тестить простым выводом в консоль
	telegramBot := service.NewTelegramBot(c.TgBot.Token)
	if telegramBot == nil {
		zlog.Logger.Warn().Msg("without telegram bot")
	}

	ampqDSN := fmt.Sprintf("amqp://%s:%s@%s/", c.RabbitMQ.User, c.RabbitMQ.Password, c.RabbitMQ.Host)
	rabbitRepo, err := repository.NewRabbitMQ(ampqDSN)
	if err != nil {
		zlog.Logger.Fatal().Msgf("failed get rabbitMQ: %v", err)
	}

	pgDSN := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Storage.DBHost, c.Storage.User, c.Storage.Password, c.Storage.DBName)
	dbStorage, err := repository.NewStorage(pgDSN)
	if err != nil {
		zlog.Logger.Fatal().Msgf("failed get DB: %v", err)
	}

	redisClient := redis.New(c.Redis.Host, c.Redis.Password, c.Redis.IdDB)

	h := handlers.NewHandler(dbStorage, rabbitRepo, redisClient)
	mux := api.Routes(&h)

	a := app.App{
		Host:     c.Server.Host,
		Handler:  mux,
		RabbitMQ: rabbitRepo,
		DB:       dbStorage,
		TgBot:    telegramBot,
		Redis:    redisClient,
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	err = a.Run(ctx)
	if err != nil {
		zlog.Logger.Fatal().Msgf("failed run app: %v", err)
	}
}
