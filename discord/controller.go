package discord

import (
	"bytes"
	"io"
	"log"
	"time"

	"github.com/Mi-lex/dgpt-bot/chat"
	"github.com/Mi-lex/dgpt-bot/config"
	"github.com/bwmarrin/discordgo"
)

var DController *Controller

type Controller struct {
	chat               *chat.Chat
	sessionClient      *discordgo.Session
	registeredCommands []*discordgo.ApplicationCommand
}

func Init(chat *chat.Chat) error {
	sessionClient, err := discordgo.New("Bot " + config.EnvConfigs.DiscordBotToken)

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

	// DController.registerCommands()
	DController.registerInteractions()

	return nil
}

func (controller *Controller) registerInteractions() {
	log.Println("Registering interaction commands...")

	controller.sessionClient.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionMessageComponent:
			if h, ok := componentsInteractionHandlers[i.MessageComponentData().CustomID]; ok {
				h(s, i)
			}
		}
	})
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

func createThreadTitle(messageContent string) string {
	if len(messageContent) < 50 {
		return messageContent
	}

	return messageContent[:50] + "..."
}

var charResponseComponents = []discordgo.MessageComponent{
	discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				CustomID: componentInteractionsEventStopResponse,
				Label:    "Stop responding",
			},
		},
	},
}

func getMockHttpResponse() io.ReadCloser {
	// Create a byte slice with your mock data
	data := []byte("This is a mock HTTP response stream")

	// Convert the byte slice to a buffer
	buffer := bytes.NewBuffer(data)

	// Return the buffer wrapped in an io.ReadCloser interface
	return io.NopCloser(buffer)
}

func (controller *Controller) messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	ch, err := s.State.Channel(m.ChannelID)

	if err != nil {
		log.Printf("Failed to get channel: %v")

		return
	}

	threadId := m.ChannelID

	// if its not a thread, then create one
	if !ch.IsThread() {
		thread, err := s.MessageThreadStartComplex(m.ChannelID, m.ID, &discordgo.ThreadStart{
			Name:             createThreadTitle(m.Content),
			Invitable:        false,
			RateLimitPerUser: 10,
		})

		if err != nil {
			log.Printf("Failed to create a thread: %v", err)

			return
		}

		threadId = thread.ID
	}

	s.ChannelTyping(threadId)

	// get response from chat
	// response, err := controller.chat.GetResponse(threadId, m.Author.ID, m.Content)

	// mocked version
	responseStream := getMockHttpResponse()

	var buffer bytes.Buffer

	var msgToEdit *discordgo.Message

	var first = true
	// Loop through the response stream and read it sequentially
	for {
		chunk := make([]byte, 5)

		n, err := responseStream.Read(chunk)

		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		buffer.Write((chunk[:n]))

		if first {
			first = false
			log.Print("Creating initial message")
			// make initial message
			msgSend := &discordgo.MessageSend{
				Content:    buffer.String(),
				Components: charResponseComponents,
			}

			msgToEdit, err = s.ChannelMessageSendComplex(threadId, msgSend)

			if err != nil {
				log.Printf("Failed to create initial message")

				break
			}
		} else {
			log.Print("Editing existing one message")
			// edit existing one
			msgEdit := discordgo.NewMessageEdit(threadId, msgToEdit.ID)
			msgEdit.SetContent(buffer.String())

			_, err = s.ChannelMessageEditComplex(msgEdit)

			if err != nil {
				log.Printf("Failed to receive %v chunk of stream", n)

				break
			}
		}

		time.Sleep(2000)
	}

	msgEdit := discordgo.NewMessageEdit(threadId, msgToEdit.ID)
	msgEdit.Components = []discordgo.MessageComponent{}
	_, err = s.ChannelMessageEditComplex(msgEdit)

	if err != nil {
		log.Printf("Failed to update final message: %v", err)

		return
	}

	// while stream is not over write add stopResponse function into controller property
	// after stream is over stopResponse should be deleted
	// if err != nil {
	// 	log.Printf("Failed to get chat response: %v", err)

	// 	return
	// }

	// if response == "" {
	// 	log.Print("Empty response")

	// 	return
	// }

	// time.Sleep(2000)

	// Edit current message on data
	// msgEdit := discordgo.NewMessageEdit(threadId, messageToEdit.ID)
	// msgEdit.SetContent("Some Content. Or event more")
	// msgEdit.Components = []discordgo.MessageComponent{}
	// _, err = s.ChannelMessageEditComplex(msgEdit)

	if err != nil {
		// If an error occurred, we failed to send a message to a channel
		// Some common causes are:
		// 1. We don't share a server with the user (not possible here).
		// 2. We opened enough DM channels quickly enough for Discord to
		//    label us as abusing the endpoint, blocking us from opening
		//    new ones.
		log.Println("error creating channel:", err)
		s.ChannelMessageSend(
			m.ChannelID,
			"Something went wrong while sending the message.",
		)
		return
	}
}
