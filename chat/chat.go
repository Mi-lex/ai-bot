package chat

import (
	"fmt"

	"github.com/Mi-lex/dgpt-bot/utils"
)

type Chat struct {
	store *Store
}

func NewChat() *Chat {
	return &Chat{
		store: &Store{
			redis: utils.RedisClient,
		},
	}
}

func createDefaultConversation(id string) *Conversation {
	return NewConversation(id, "gpt-3.5-turbo", 0.2)
}

func (chat *Chat) GetResponse(conversationId string, userId string, message string) (response string, err error) {
	conversation, err := chat.store.GetConversation(conversationId)

	if err != nil {
		fmt.Printf("Failed to get conversation with id: %s, err: %s", conversationId, err)

		return "", err
	}

	if conversation == nil {
		conversation = createDefaultConversation(conversationId)
	}

	conversation.AddUserContext(message)

	chat.store.SetConversation(conversation)

	return "its done", nil
}
