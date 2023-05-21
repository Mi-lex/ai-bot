package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Mi-lex/dgpt-bot/commands_example"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	commands_example.Echo()

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	botToken := os.Getenv("DISCORD_BOT_TOKEN")
	// Create a new Discord session using the provided bot token.
	discordClient, err := discordgo.New("Bot " + botToken)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	discordClient.AddHandler(messageCreate)

	discordClient.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commands_example.CommandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	log.Println("Adding commands...")
	// Just like the ping pong example, we only care about receiving message
	// events in this example.
	// dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = discordClient.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands_example.Commands))

	for i, commandOption := range commands_example.Commands {
		cmd, err := discordClient.ApplicationCommandCreate(discordClient.State.User.ID, "", commandOption)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", commandOption.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer discordClient.Close()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop

	log.Println("Removing commands...")

	for _, registeredCommand := range registeredCommands {
		err := discordClient.ApplicationCommandDelete(discordClient.State.User.ID, "", registeredCommand.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", registeredCommand.Name, err)
		}
	}
	// Cleanly close down the Discord session.
	log.Println("Gracefully shutting down.")
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
//
// It is called whenever a message is created but only when it's sent through a
// server as we did not request IntentsDirectMessages.
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
