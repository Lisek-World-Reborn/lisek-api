package channels

import (
	"context"
	"os"

	"github.com/Lisek-World-Reborn/lisek-api/config"
	"github.com/Lisek-World-Reborn/lisek-api/logger"
	"github.com/go-redis/redis/v9"
)

var RedisConnection *redis.Client

func Init() {
	logger.Info("Redis channels initialized")

	RedisConnection = redis.NewClient(&redis.Options{
		Addr:     config.LoadedConfiguration.Redis.Address,
		Password: config.LoadedConfiguration.Redis.Password,
		DB:       0,
	})

	_, err := RedisConnection.Ping(context.Background()).Result()

	if err != nil {
		logger.Fatal(err.Error())
		os.Exit(0)
	}

	logger.Info("Redis ping - OK")
}
