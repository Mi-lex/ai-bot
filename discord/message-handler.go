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

// chunk size should be smaller than discord message max len
const chunkMaxLen = 100
const discordMessageMaxLen = 2000

var currentChatResponses = make(map[string]func())

func (controller *Controller) messageHandler(session *discordGoLib.Session, message *discordGoLib.MessageCreate) {
	// ignore all messages created by the bot itself
	if message.Author.ID == session.State.User.ID {
		return
	}

	channel, err := session.State.Channel(message.ChannelID)

	if err != nil {
		handleDiscordChatError(session, message.ChannelID, fmt.Sprintf("Failed to get a channel %v:%v", message.ChannelID, err))

		return
	}

	threadId := message.ChannelID

	_, isResponding := currentChatResponses[threadId]

	// if chat in the same thread already responding
	if isResponding {
		return
	}

	// if its not a thread, then create one
	if !channel.IsThread() {
		threadId, err = createThread(session, message)

		if err != nil {
			handleDiscordChatError(session, message.ChannelID, fmt.Sprintf("Failed to create a thread: %v", err))

			return
		}
	}

	// assign empty function for now
	currentChatResponses[threadId] = func() {}

	// make sure we always remove response
	defer delete(currentChatResponses, threadId)

	controller.writeAnswer(session, threadId, message)
}

func createThread(session *discordGoLib.Session, message *discordGoLib.MessageCreate) (string, error) {
	thread, err := session.MessageThreadStartComplex(message.ChannelID, message.ID, &discordGoLib.ThreadStart{
		Name:             createThreadTitle(message.Content),
		Invitable:        false,
		RateLimitPerUser: 10,
	})

	if err != nil {
		return "", err
	}

	return thread.ID, nil
}

func (controller *Controller) writeAnswer(session *discordGoLib.Session, threadId string, message *discordGoLib.MessageCreate) {
	var chatResponseMessage *discordGoLib.Message

	/*
	* we send discord messages by chunks
	* to reduce request amount during chat response
	 */
	var chunk = ""
	var chatResponse = ""
	var err error = nil

	session.ChannelTyping(threadId)
	// get response from chat
	err = controller.chat.GetStreamResponse(threadId, message.Author.ID, message.Content, func(data string, stop func()) {
		// if response is bigger than discord message max len
		if len(data)+len(chatResponse) >= discordMessageMaxLen {
			// assuming that chunk is always smaller than discordMessageMaxLen
			chatResponse = chunk + data

			// remove button for previous msg
			msgToEdit := discordGoLib.NewMessageEdit(threadId, chatResponseMessage.ID)
			msgToEdit.Components = []discordGoLib.MessageComponent{}
			_, err = session.ChannelMessageEditComplex(msgToEdit)

			if err != nil {
				handleDiscordChatError(session, threadId, fmt.Sprintf("Failed to edit previous message: %v", err))
			}
		} else {
			chatResponse += data
		}

		// if no data left
		if data == "" {
			// if it's last message
			if chatResponseMessage != nil {
				// we edit existing one, removing "button" component
				msgToEdit := discordGoLib.NewMessageEdit(threadId, chatResponseMessage.ID)
				msgToEdit.Components = []discordGoLib.MessageComponent{}
				msgToEdit.SetContent(chatResponse + chunk)
				_, err := session.ChannelMessageEditComplex(msgToEdit)

				if err != nil {
					handleDiscordChatError(session, threadId, fmt.Sprintf("Failed to edit final message: %v", err))
				}
				// if full response is bigger than chunkMaxLen and it's first response
			} else if chatResponse != "" {
				// we chat send simple message
				chatResponseMessage, err = session.ChannelMessageSend(threadId, chatResponse)

				if err != nil {
					handleDiscordChatError(session, threadId, fmt.Sprintf("Failed to send simple message: %v", err))
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

				chatResponseMessage, err = session.ChannelMessageSendComplex(threadId, msgSend)

				if err != nil {
					handleDiscordChatError(session, threadId, fmt.Sprintf("Failed to create an initial message: %v", err))
				}
			} else {
				// edit existing one
				msgToEdit := discordGoLib.NewMessageEdit(threadId, chatResponseMessage.ID)
				msgToEdit.SetContent(chatResponse)

				_, err := session.ChannelMessageEditComplex(msgToEdit)

				if err != nil {
					handleDiscordChatError(session, threadId, fmt.Sprintf("Failed to edit existing message: %v", err))
				}
			}

			chunk = ""
		}
	})

	if err != nil {
		handleDiscordChatError(session, message.ChannelID, fmt.Sprintf("Failed to get chat stream: %v", err))
	}
}

type onData func(content string, stop func())

func handleDiscordChatError(session *discordGoLib.Session, channelId string, errorMessage string) {
	log.Println(errorMessage)

	session.ChannelMessageSend(channelId, errorMessage)
}
