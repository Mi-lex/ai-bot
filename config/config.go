package config

import (
	"log"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

var EnvConfigs *envConfigs

func InitEnvConfigs() {
	EnvConfigs = loadEnvVariables()
}

type envConfigs struct {
	DiscordBotToken        string
	DiscordGuildId         string
	OpenApiSecretKey       string
	OpenApiAvailableModels []string
	RedisAddr              string
	RedisPass              string
}

const (
	OpenApiGpt3_5Turbo    = "gpt-3.5-turbo"
	OpenApiGpt3_5Turbo16k = "gpt-3.5-turbo-16k"
	OpenApiGpt4           = "gpt-4"
	OpenApiGpt4Turbo      = "gpt-4-turbo"
	OpenApiGpt4o          = "gpt-4o"
)

const DEFAULT_OPENAI_MODEL = OpenApiGpt3_5Turbo

var DEFAULT_OPENAI_MODELS = []string{OpenApiGpt3_5Turbo, OpenApiGpt3_5Turbo16k, OpenApiGpt4, OpenApiGpt4Turbo, OpenApiGpt4o}

var defaultConfigs = map[string]interface{}{
	"OPENAI_AVAILABLE_MODELS": DEFAULT_OPENAI_MODELS,
}

func loadEnvVariables() (config *envConfigs) {
	var k = koanf.New(".")

	k.Load(confmap.Provider(defaultConfigs, "."), nil)

	if err := k.Load(file.Provider(".env"), dotenv.Parser()); err != nil {
		log.Println("Error loading config:", err)
	}

	k.Load(env.Provider("", ".", func(s string) string {
		return s
	}), nil)

	return &envConfigs{
		DiscordBotToken:        k.String("DISCORD_BOT_TOKEN"),
		DiscordGuildId:         k.String("DISCORD_GUILD_ID"),
		OpenApiSecretKey:       k.String("OPENAI_SECRET_KEY"),
		OpenApiAvailableModels: k.Strings("OPENAI_AVAILABLE_MODELS"),
		RedisAddr:              k.String("REDIS_ADDR"),
		RedisPass:              k.String("REDIS_PASS"),
	}
}
