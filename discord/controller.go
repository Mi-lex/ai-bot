package discord

import (
	"bytes"
	"fmt"
	"io"
	"log"

	"github.com/Mi-lex/dgpt-bot/chat"
	"github.com/Mi-lex/dgpt-bot/config"
	discordGoLib "github.com/bwmarrin/discordgo"
)

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

	// DController.registerCommands()
	DController.registerInteractions()

	return nil
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

var charResponseComponents = []discordGoLib.MessageComponent{
	discordGoLib.ActionsRow{
		Components: []discordGoLib.MessageComponent{
			discordGoLib.Button{
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

var chunkMaxLen = 100

func (controller *Controller) messageHandler(s *discordGoLib.Session, m *discordGoLib.MessageCreate) {
	// ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	ch, err := s.State.Channel(m.ChannelID)

	if err != nil {
		log.Println("Failed to get channel:", err)

		return
	}

	threadId := m.ChannelID

	// if its not a thread, then create one
	if !ch.IsThread() {
		thread, err := s.MessageThreadStartComplex(m.ChannelID, m.ID, &discordGoLib.ThreadStart{
			Name:             createThreadTitle(m.Content),
			Invitable:        false,
			RateLimitPerUser: 10,
		})

		if err != nil {
			log.Println("Failed to create a thread:", err)

			return
		}

		threadId = thread.ID
	}

	var assistantResponseMessage *discordGoLib.Message

	// we send discord messages by chunks
	// to reduce request amount during chat response
	var chunk = ""
	var chatResponse = ""

	s.ChannelTyping(threadId)
	// get response from chat
	err = controller.chat.GetStreamResponse(threadId, m.Author.ID, m.Content, func(data string) {
		chatResponse += data

		// if no data left
		if data == "" {
			// if it's last message
			if assistantResponseMessage != nil {
				// we edit existing one, removing "button" component
				msgToEdit := discordGoLib.NewMessageEdit(threadId, assistantResponseMessage.ID)
				msgToEdit.Components = []discordGoLib.MessageComponent{}
				msgToEdit.SetContent(chatResponse)
				_, err = s.ChannelMessageEditComplex(msgToEdit)

				if err != nil {
					log.Println("Failed to edit final message:", err)
				}
				// if full response is bigger than chunkMaxLen and it's first response
			} else if chatResponse != "" {
				// we chat send simple message
				assistantResponseMessage, err = s.ChannelMessageSend(threadId, chatResponse)

				if err != nil {
					log.Println("Failed to send simple message:", err)
				}
			}

			return
		}

		chunk += data

		if len(chunk) >= chunkMaxLen {
			// first message
			if chatResponse == chunk {
				// make initial message
				msgSend := &discordGoLib.MessageSend{
					Content:    chatResponse,
					Components: charResponseComponents,
				}

				assistantResponseMessage, err = s.ChannelMessageSendComplex(threadId, msgSend)

				if err != nil {
					log.Println("Failed to create an editable initial message:")
				}
			} else {
				log.Println("Editing the existing message")
				// edit existing one
				msgToEdit := discordGoLib.NewMessageEdit(threadId, assistantResponseMessage.ID)
				msgToEdit.SetContent(chatResponse)

				_, err = s.ChannelMessageEditComplex(msgToEdit)

				if err != nil {
					log.Println("Failed to edit existing message:", err)
				}
			}

			chunk = ""
		}
	})

	if err != nil {
		errorMessage := fmt.Errorf("Failed to get chat stream: %v").Error()

		s.ChannelMessageSend(
			m.ChannelID,
			errorMessage,
		)

		return
	}
}
