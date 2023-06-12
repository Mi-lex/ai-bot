package utils

import (
	"fmt"

	"github.com/Mi-lex/dgpt-bot/config"
	redis "github.com/go-redis/redis/v8"
	"github.com/nitishm/go-rejson/v4"
)

type Redis struct {
	client     *redis.Client
	jsonClient *rejson.Handler
}

func (r *Redis) ToBytes(jsonStr string) []byte {
	jsonBytes := []byte(jsonStr)

	return jsonBytes
}

func (r *Redis) GetJson(key string) (data interface{}, err error) {
	res, err := r.jsonClient.JSONGet(key, ".")

	if err != nil {
		return nil, fmt.Errorf("Failed to JSONGet from redis: %w", err)
	}

	return res, nil
}

func (r *Redis) SetJson(key string, data interface{}) (res interface{}, err error) {
	res, err = r.jsonClient.JSONSet(key, ".", data)

	if err != nil {
		return nil, fmt.Errorf("Failed to JSONGet from redis: %w", err)
	}

	return res, nil
}

var RedisClient *Redis

func SetupRedis() error {
	if config.EnvConfigs.RedisAddr == "" {
		return fmt.Errorf("REDIS_ADDR is not defined")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     config.EnvConfigs.RedisAddr,
		Password: config.EnvConfigs.RedisPass,
		DB:       0,
	})

	jsonClient := rejson.NewReJSONHandler()
	jsonClient.SetGoRedisClient(client)

	RedisClient = &Redis{
		client:     client,
		jsonClient: jsonClient,
	}

	return nil
}
