package discord

import (
	"fmt"
	"log"

	discordGoLib "github.com/bwmarrin/discordgo"
)

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

var chunkMaxLen = 100

var currentChatResponses = make(map[string]func())

func (controller *Controller) messageHandler(s *discordGoLib.Session, m *discordGoLib.MessageCreate) {
	// ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	ch, err := s.State.Channel(m.ChannelID)

	if err != nil {
		log.Println("Failed to get a channel %v:%v", m.ChannelID, err)

		return
	}

	threadId := m.ChannelID

	_, isResponding := currentChatResponses[threadId]

	// if chat in the same thread already responding
	if isResponding {
		return
	}

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

	// assign empty function for now
	currentChatResponses[threadId] = func() {}

	// make sure we always remove response
	defer delete(currentChatResponses, threadId)

	var chatResponseMessage *discordGoLib.Message

	// we send discord messages by chunks
	// to reduce request amount during chat response
	var chunk = ""
	var chatResponse = ""

	s.ChannelTyping(threadId)
	// get response from chat
	err = controller.chat.GetStreamResponse(threadId, m.Author.ID, m.Content, func(data string, stop func()) {
		chatResponse += data

		// if no data left
		if data == "" {
			// if it's last message
			if chatResponseMessage != nil {
				// we edit existing one, removing "button" component
				msgToEdit := discordGoLib.NewMessageEdit(threadId, chatResponseMessage.ID)
				msgToEdit.Components = []discordGoLib.MessageComponent{}
				msgToEdit.SetContent(chatResponse)
				_, err = s.ChannelMessageEditComplex(msgToEdit)

				if err != nil {
					log.Println("Failed to edit final message:", err)
				}
				// if full response is bigger than chunkMaxLen and it's first response
			} else if chatResponse != "" {
				// we chat send simple message
				chatResponseMessage, err = s.ChannelMessageSend(threadId, chatResponse)

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
				currentChatResponses[threadId] = stop

				// make initial message
				msgSend := &discordGoLib.MessageSend{
					Content:    chatResponse,
					Components: charResponseComponents,
				}

				chatResponseMessage, err = s.ChannelMessageSendComplex(threadId, msgSend)

				if err != nil {
					log.Println("Failed to create an initial message:", err)
				}
			} else {
				// edit existing one
				msgToEdit := discordGoLib.NewMessageEdit(threadId, chatResponseMessage.ID)
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
