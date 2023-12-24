package discord

import (
	"fmt"

	discordGoLib "github.com/bwmarrin/discordgo"
)

type ResponseMessage struct {
	content        string
	threadId       string
	discordMessage *discordGoLib.Message
	session        *discordGoLib.Session
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

const maxMessageLen = 2000

func NewResponseMessage(session *discordGoLib.Session, threadId string) *ResponseMessage {
	return &ResponseMessage{
		session:  session,
		threadId: threadId,
	}
}

func (responseMessage *ResponseMessage) getCurrentContent() string {
	if responseMessage.discordMessage != nil {
		return responseMessage.discordMessage.Content
	}

	return ""
}

func (responseMessage *ResponseMessage) WillOverFlow(dataToAppend string) bool {
	return len(responseMessage.getCurrentContent())+len(dataToAppend) > maxMessageLen
}

func (responseMessage *ResponseMessage) edit(newContent string) error {
	msgToEdit := discordGoLib.NewMessageEdit(responseMessage.threadId, responseMessage.discordMessage.ID)
	msgToEdit.SetContent(newContent)

	msg, error := responseMessage.session.ChannelMessageEditComplex(msgToEdit)

	if error != nil {
		return fmt.Errorf("failed to edit previous message: %v", error)
	}

	responseMessage.discordMessage = msg

	return nil
}

/**
* The function either will send will just send message
* or update existing one with given content and will
* remove buttons
**/
func (responseMessage *ResponseMessage) Finalize(content string) error {
	var err error = nil

	if responseMessage.discordMessage == nil {
		if content == "" {
			return nil
		}

		_, err = responseMessage.session.ChannelMessageSend(responseMessage.threadId, content)

		return err
	}

	msgToEdit := discordGoLib.NewMessageEdit(responseMessage.threadId, responseMessage.discordMessage.ID)
	msgToEdit.Components = []discordGoLib.MessageComponent{}

	if content != "" {
		msgToEdit.SetContent(content)
	}

	_, error := responseMessage.session.ChannelMessageEditComplex(msgToEdit)

	if error != nil {
		return fmt.Errorf("failed to edit previous message: %v", error)
	}

	responseMessage.discordMessage = nil

	return nil
}

func (responseMessage *ResponseMessage) Send(content string) error {
	var err error = nil

	// if message exist we just edit it
	if responseMessage.discordMessage != nil {
		err = responseMessage.edit(content)
	} else {
		// otherwise we create new one
		msgToSend := &discordGoLib.MessageSend{
			Content:    content,
			Components: charResponseComponents,
		}

		responseMessage.discordMessage, err = responseMessage.session.ChannelMessageSendComplex(responseMessage.threadId, msgToSend)
	}

	if err != nil {
		responseMessage.content = content
	}

	return err
}
