package config

import (
	"log"

	"github.com/spf13/viper"
)

// Initilize this variable to access the env values
var EnvConfigs *envConfigs

// We will call this in main.go to load the env variables
func InitEnvConfigs() {
	EnvConfigs = loadEnvVariables()
}

// struct to map env values
type envConfigs struct {
	DiscordBotToken  string `mapstructure:"DISCORD_BOT_TOKEN"`
	OpenApiSecretKey string `mapstructure:"OPENAI_SECRET_KEY"`
	RedisAddr        string `mapstructure:"REDIS_ADDR"`
	RedisPass        string `mapstructure:"REDIS_PASS"`
	Bruh             string `mapstructure:"BRUH"`
}

func loadEnvVariables() (config *envConfigs) {
	viper.AutomaticEnv()

	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	// Viper reads all the variables from env file and log error if any found
	if err := viper.ReadInConfig(); err != nil {
		log.Print("Failed to load .env. Using environmental variables")
	}

	// Viper unmarshals the loaded env varialbes into the struct
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal(err)
	}

	return
}
