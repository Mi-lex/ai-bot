package discord

import (
	"log"

	"github.com/Mi-lex/dgpt-bot/chat"
	"github.com/Mi-lex/dgpt-bot/config"
	discordGoLib "github.com/bwmarrin/discordgo"
)

// TODO there is no need in this variable
var DController *Controller

type Controller struct {
	chat               *chat.Chat
	sessionClient      *discordGoLib.Session
	registeredCommands []*discordGoLib.ApplicationCommand
}

func Init(chat *chat.Chat) error {
	sessionClient, err := discordGoLib.New("Bot " + config.EnvConfigs.DiscordBotToken)

	if err != nil {
		log.Fatal("error creating Discord session,", err)

		return err
	}

	DController = &Controller{
		sessionClient: sessionClient,
		chat:          chat,
	}

	DController.sessionClient.AddHandler(DController.messageHandler)

	err = DController.sessionClient.Open()

	if err != nil {
		log.Println("error opening connection,", err)
		return err
	}

	DController.registerCommands()
	DController.registerInteractions()

	return nil
}

func (controller *Controller) registerCommands() {
	log.Println("Registering commands...")

	setModeHandler := createSetModelHandler(controller.chat.SetModel)

	controller.sessionClient.AddHandler(func(s *discordGoLib.Session, i *discordGoLib.InteractionCreate) {
		switch i.Type {
		case discordGoLib.InteractionApplicationCommand:
			if i.ApplicationCommandData().Name == SET_MODEL_COMMAND {
				setModeHandler(s, i)
			}
		}
	})

	cmd, err := controller.sessionClient.ApplicationCommandCreate(controller.sessionClient.State.User.ID, config.EnvConfigs.DiscordGuildId, createSetModelCommand(config.EnvConfigs.OpenApiAvailableModels))

	if err != nil {
		log.Panicf("Cannot create '%v' command: %v", "Set model", err)
	}

	controller.registeredCommands = append(controller.registeredCommands, cmd)
}

func (controller *Controller) registerInteractions() {
	log.Println("Registering interaction commands...")

	controller.sessionClient.AddHandler(func(s *discordGoLib.Session, i *discordGoLib.InteractionCreate) {
		switch i.Type {
		case discordGoLib.InteractionMessageComponent:
			if h, ok := componentsInteractionHandlers[i.MessageComponentData().CustomID]; ok {
				h(s, i)
			}
		}
	})
}

func (controller *Controller) unregisterCommands() {
	log.Println("Removing commands...")

	for _, registeredCommand := range controller.registeredCommands {
		err := controller.sessionClient.ApplicationCommandDelete(controller.sessionClient.State.User.ID, config.EnvConfigs.DiscordGuildId, registeredCommand.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", registeredCommand.Name, err)
		}
	}
}

func (controller *Controller) Close() {
	// Cleanly close down the Discord session.
	log.Println("Gracefully shutting down.")

	controller.unregisterCommands()
	controller.sessionClient.Close()
}
