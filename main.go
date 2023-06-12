package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Mi-lex/dgpt-bot/chat"
	"github.com/Mi-lex/dgpt-bot/config"
	"github.com/Mi-lex/dgpt-bot/discord"
	"github.com/Mi-lex/dgpt-bot/utils"

	openai "github.com/sashabaranov/go-openai"
)

func main() {
	config.InitEnvConfigs()

	err := utils.SetupRedis()

	if err != nil {
		log.Fatal("error setting up Redis,", err)
	}

	openAiClient := openai.NewClient(config.EnvConfigs.OpenApiSecretKey)

	chat := chat.NewChat(openAiClient)

	err = discord.Init(chat)

	if err != nil {
		log.Fatal("error creating Discord session,", err)

		return
	}

	defer discord.DController.Close()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop
}
