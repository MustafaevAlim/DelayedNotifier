package config

import "github.com/wb-go/wbf/config"

type ConfigApp struct {
	Server   ServerConfig
	RabbitMQ RabbitMQConfig
	Storage  StorageConfig
	TgBot    TgBotConfig
	Redis    RedisConfig
}

type ServerConfig struct {
	Host string
}

type TgBotConfig struct {
	Token string
}

type RabbitMQConfig struct {
	Host     string
	User     string
	Password string
}

type StorageConfig struct {
	DBHost   string
	User     string
	Password string
	DBName   string
}

type RedisConfig struct {
	Host     string
	Password string
	IdDB     int
}

func NewConfig() (*ConfigApp, error) {
	c := config.New()
	err := c.Load(".env")
	if err != nil {
		return nil, err
	}

	cfg := ConfigApp{
		Server: ServerConfig{
			Host: c.GetString("HOST"),
		},
		RabbitMQ: RabbitMQConfig{
			Host:     c.GetString("RABBITMQ_HOST"),
			User:     c.GetString("RABBITMQ_USER"),
			Password: c.GetString("RABBITMQ_PASSWORD"),
		},
		Storage: StorageConfig{
			DBHost:   c.GetString("POSTGRES_HOST"),
			User:     c.GetString("POSTGRES_USER"),
			Password: c.GetString("POSTGRES_PASSWORD"),
			DBName:   c.GetString("POSTGRES_DB"),
		},
		TgBot: TgBotConfig{
			Token: c.GetString("TOKEN_TG_BOT"),
		},
		Redis: RedisConfig{
			Host:     c.GetString("REDIS_HOST"),
			Password: c.GetString("REDIS_PASSWORD"),
			IdDB:     c.GetInt("REDIS_DB_INT"),
		},
	}

	return &cfg, nil

}
