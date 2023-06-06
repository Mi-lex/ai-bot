package discord

import (
	"fmt"
	"log"

	"github.com/Mi-lex/dgpt-bot/commands_example"
	"github.com/Mi-lex/dgpt-bot/config"
	"github.com/bwmarrin/discordgo"
)

var DController *Controller

type Controller struct {
	sessionClient *discordgo.Session
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

	DController.sessionClient.AddHandler(messageCreate)

	DController.sessionClient.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commands_example.CommandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	err = DController.sessionClient.Open()

	if err != nil {
		fmt.Println("error opening connection,", err)
		return err
	}

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands_example.Commands))

	log.Println("Registering commands...")
	for i, commandOption := range commands_example.Commands {
		cmd, err := DController.sessionClient.ApplicationCommandCreate(DController.sessionClient.State.User.ID, "", commandOption)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", commandOption.Name, err)
		}
		registeredCommands[i] = cmd
	}

	return nil
}

func (controller *Controller) Close() {
	controller.sessionClient.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// threadMessages := s.thre
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	fmt.Print("Message created\n")

	// m.Flags
	fmt.Printf("Is thread %v \n", m.Flags)

	fmt.Printf("Channel id: %v\n", m.ChannelID)

	ch, _ := s.State.Channel(m.ChannelID)

	fmt.Printf("Is thread %v \n", ch.IsThread())
	// if m.Thread.IsThread() {
	// 	fmt.Printf("Messages %v", m.Thread.Messages)
	// }

	fmt.Printf(m.Author.ID + " \n")
	fmt.Printf(s.State.User.ID + " \n")

	if m.Author.ID == s.State.User.ID {
		fmt.Printf("Written by bot\n")
		return
	}

	fmt.Printf("Here's the message: \n")
	fmt.Printf(m.Message.Content)
	// fmt.Printf(m.UnmarshalJSON([]byte(m.Content)))

	// In this example, we only care about messages that are "ping".
	if m.Content != "ping" {
		return
	}

	// We create the private channel with the user who sent the message.
	// channel, err := s.UserChannelCreate(m.Author.ID)
	_, err := s.ChannelMessageSend(m.ChannelID, "pong")

	// m.Thread.IsThread()
	// m.Thread.Messages

	if err != nil {
		// If an error occurred, we failed to create the channel.
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
