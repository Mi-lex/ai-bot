package config

import (
	"log"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var EnvConfigs *envConfigs

func InitEnvConfigs() {
	EnvConfigs = loadEnvVariables()
}

type envConfigs struct {
	DiscordBotToken  string
	OpenApiSecretKey string
	RedisAddr        string
	RedisPass        string
}

func loadEnvVariables() (config *envConfigs) {
	var k = koanf.New(".")

	if err := k.Load(file.Provider(".env"), dotenv.Parser()); err != nil {
		log.Println("Error loading config:", err)
	}

	k.Load(env.Provider("", ".", func(s string) string {
		return s
	}), nil)

	return &envConfigs{
		DiscordBotToken:  k.String("DISCORD_BOT_TOKEN"),
		OpenApiSecretKey: k.String("OPENAI_SECRET_KEY"),
		RedisAddr:        k.String("REDIS_ADDR"),
		RedisPass:        k.String("REDIS_PASS"),
	}
}
