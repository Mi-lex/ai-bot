package chat

import (
	"context"
	"fmt"
	"log"

	"github.com/Mi-lex/dgpt-bot/utils"
	openai "github.com/sashabaranov/go-openai"
)

type Chat struct {
	store        *Store
	openAiClient *openai.Client
}

func NewChat(openAiCLient *openai.Client) *Chat {
	return &Chat{
		store: &Store{
			redis: utils.RedisClient,
		},
		openAiClient: openAiCLient,
	}
}

func createDefaultConversation(id string) *Conversation {
	return NewConversation(id, "gpt-3.5-turbo", 0.2)
}

func (chat *Chat) createChatCompletionStream(conversation *Conversation) (*openai.ChatCompletionStream, error) {
	var messages = make([]openai.ChatCompletionMessage, len(conversation.ContextList))

	for i, context := range conversation.ContextList {
		messages[i] = openai.ChatCompletionMessage{
			Role:    context.Role,
			Content: context.Content,
		}
	}

	ctx := context.Background()
	request := openai.ChatCompletionRequest{
		Model:    openai.GPT3Dot5Turbo,
		Messages: messages,
		Stream:   true,
	}

	stream, err := chat.openAiClient.CreateChatCompletionStream(ctx, request)

	if err != nil {
		log.Print("Failed to create stream chat completion: %v", err)

		return nil, fmt.Errorf("Failed to create stream chat completion: %v", err)
	}

	return stream, nil
}

func (chat *Chat) createChatCompletion(conversation *Conversation) (*openai.ChatCompletionMessage, error) {
	var messages = make([]openai.ChatCompletionMessage, len(conversation.ContextList))

	for i, context := range conversation.ContextList {
		messages[i] = openai.ChatCompletionMessage{
			Role:    context.Role,
			Content: context.Content,
		}
	}

	ctx := context.Background()
	request := openai.ChatCompletionRequest{
		Model:    openai.GPT3Dot5Turbo,
		Messages: messages,
	}

	resp, err := chat.openAiClient.CreateChatCompletion(
		ctx,
		request,
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to create chat completion: %v", err)
	}

	return &resp.Choices[0].Message, nil
}

func (chat *Chat) GetResponse(conversationId string, userId string, message string) (response string, err error) {
	conversation, err := chat.store.GetConversation(conversationId)

	if err != nil {
		return "", fmt.Errorf("Failed to get conversation with id: %s: %v", conversationId, err)
	}

	if conversation == nil {
		conversation = createDefaultConversation(conversationId)
	}

	conversation.AddUserContext(message)

	chatCompletionMessage, err := chat.createChatCompletion(conversation)

	if err != nil {
		return "", err
	}

	conversation.AddAssistantContent(chatCompletionMessage.Content)
	chat.store.SetConversation(conversation)

	return chatCompletionMessage.Content, nil
}
