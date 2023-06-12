package discord

import (
	"fmt"
	"log"
	"time"

	"github.com/Mi-lex/dgpt-bot/config"
	"github.com/bwmarrin/discordgo"
)

var DController *Controller

type Controller struct {
	sessionClient      *discordgo.Session
	registeredCommands []*discordgo.ApplicationCommand
}

func Init() error {
	sessionClient, err := discordgo.New("Bot " + config.EnvConfigs.DiscordBotToken)

	if err != nil {
		log.Fatal("error creating Discord session,", err)

		return err
	}

	DController = &Controller{
		sessionClient: sessionClient,
	}

	DController.sessionClient.AddHandler(userMessageHandler)

	DController.sessionClient.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := CommandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	err = DController.sessionClient.Open()

	if err != nil {
		fmt.Println("error opening connection,", err)
		return err
	}

	DController.registerCommands()

	return nil
}

func (controller *Controller) registerCommands() {
	controller.registeredCommands = make([]*discordgo.ApplicationCommand, len(Commands))

	log.Println("Registering commands...")
	for i, commandOption := range Commands {
		cmd, err := DController.sessionClient.ApplicationCommandCreate(DController.sessionClient.State.User.ID, "", commandOption)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", commandOption.Name, err)
		}

		controller.registeredCommands[i] = cmd
	}
}

func (controller *Controller) unregisterCommands() {
	log.Println("Removing commands...")

	for _, registeredCommand := range controller.registeredCommands {
		err := controller.sessionClient.ApplicationCommandDelete(controller.sessionClient.State.User.ID, "", registeredCommand.ID)
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

func userMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	ch, err := s.State.Channel(m.ChannelID)

	if err != nil {
		fmt.Printf("Error getting channel %v", err)

		return
	}

	threadId := m.ChannelID

	// if its not a thread, then create one
	if !ch.IsThread() {
		thread, err := s.MessageThreadStartComplex(m.ChannelID, m.ID, &discordgo.ThreadStart{
			Name:             "Thread",
			Invitable:        false,
			RateLimitPerUser: 10,
		})

		if err != nil {
			fmt.Printf("Error creating thread %v", err)

			return
		}

		threadId = thread.ID
	}

	s.ChannelTyping(threadId)

	// Sleep for 2 seconds to simulate a long-running task.
	time.Sleep(2 * 1e9)

	_, err = s.ChannelMessageSend(threadId, "pong")

	if err != nil {
		// If an error occurred, we failed to send a message to a channel
		//
		// Some common causes are:
		// 1. We don't share a server with the user (not possible here).
		// 2. We opened enough DM channels quickly enough for Discord to
		//    label us as abusing the endpoint, blocking us from opening
		//    new ones.
		fmt.Println("error creating channel:", err)
		s.ChannelMessageSend(
			m.ChannelID,
			"Something went wrong while sending the message.",
		)
		return
	}
}
