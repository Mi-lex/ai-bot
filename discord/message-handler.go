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

// chunk size should be smaller than discord message max len
const chunkMaxLen = 100

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

	responseMessage := NewResponseMessage(session, threadId)

	session.ChannelTyping(threadId)

	err = controller.chat.GetStreamResponse(threadId, message.Author.ID, message.Content, func(data string, stop func()) {
		// if response is bigger than discord message max len
		if responseMessage.WillOverFlow(data) {
			err = responseMessage.Finalize("")

			// reset response
			chatResponse = chunk + data
		} else {
			chatResponse += data
		}

		// if no data left
		if data == "" {
			var contentToSend = chatResponse

			// if it's last message
			if chatResponseMessage != nil {
				contentToSend += chunk

				err = responseMessage.Finalize(contentToSend)
			} else if chatResponse != "" {
				contentToSend = chatResponse

				err = responseMessage.Finalize(contentToSend)
			}

			return
		}

		chunk += data

		if len(chunk) >= chunkMaxLen {
			// if first message
			if chatResponse == chunk {
				currentChatResponses[threadId] = stop
			}

			err = responseMessage.Send(chatResponse)

			chunk = ""
		}
	})

	if err != nil {
		handleDiscordChatError(session, message.ChannelID, fmt.Sprintf("Failed to get chat stream: %v", err))
	}
}

func handleDiscordChatError(session *discordGoLib.Session, channelId string, errorMessage string) {
	log.Println(errorMessage)

	session.ChannelMessageSend(channelId, errorMessage)
}
