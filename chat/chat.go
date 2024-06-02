package chat

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"

	"github.com/Mi-lex/dgpt-bot/utils"
	openAiLib "github.com/sashabaranov/go-openai"
)

type Chat struct {
	store        *Store
	openAiClient *openAiLib.Client
}

func NewChat(openAiCLient *openAiLib.Client) *Chat {
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

func (chat *Chat) createChatCompletionStream(conversation *Conversation) (*openAiLib.ChatCompletionStream, error) {
	var messages = make([]openAiLib.ChatCompletionMessage, len(conversation.ContextList))

	for i, context := range conversation.ContextList {
		messages[i] = openAiLib.ChatCompletionMessage{
			Role:    context.Role,
			Content: context.Content,
		}
	}

	ctx := context.Background()

	request := openAiLib.ChatCompletionRequest{
		Model:    conversation.Model,
		Messages: messages,
		Stream:   true,
	}

	stream, err := chat.openAiClient.CreateChatCompletionStream(ctx, request)

	if err != nil {
		log.Println("Failed to create stream chat completion:", err)

		return nil, fmt.Errorf("Failed to create stream chat completion: %v", err)
	}

	return stream, nil
}

func (chat *Chat) createChatCompletion(conversation *Conversation) (*openAiLib.ChatCompletionMessage, error) {
	var messages = make([]openAiLib.ChatCompletionMessage, len(conversation.ContextList))

	for i, context := range conversation.ContextList {
		messages[i] = openAiLib.ChatCompletionMessage{
			Role:    context.Role,
			Content: context.Content,
		}
	}

	ctx := context.Background()
	request := openAiLib.ChatCompletionRequest{
		Model:    conversation.Model,
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

type onData func(content string, stop func())

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (chat *Chat) MockStreamResponse(conversationId string, userId string, message string, onData onData) (err error) {
	var fakeTextLen = 2100

	var msg = RandStringRunes(fakeTextLen)

	dataSize := 10

	for i := 0; i < fakeTextLen; i += dataSize {
		if i+dataSize > fakeTextLen {
			dataSize = fakeTextLen - i
		}

		fakeText := msg[i : i+dataSize]

		onData(fakeText, func() {})
	}

	// we finish with empty message just like chatgpt does
	onData("", func() {})

	return nil
}

func (chat *Chat) GetStreamResponse(conversationId string, userId string, message string, onData onData) (err error) {
	conversation, err := chat.store.GetConversation(conversationId)

	if err != nil {
		return fmt.Errorf("Failed to get conversation with id: %s: %v\n", conversationId, err)
	}

	if conversation == nil {
		conversation = createDefaultConversation(conversationId)
	}

	conversation.AddUserContext(message)

	responseStream, err := chat.createChatCompletionStream(conversation)

	if err != nil {
		return fmt.Errorf("ChatCompletionStream error: %v\n", err)
	}

	defer responseStream.Close()

	var fullContent = ""

	for {
		streamResponse, err := responseStream.Recv()

		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			if err.Error() == "http2: response body closed" {
				return nil
			}

			return err
		}

		content := streamResponse.Choices[0].Delta.Content
		fullContent += content

		onData(content, func() {
			responseStream.Close()
		})
	}

	conversation.AddAssistantContent(fullContent)

	chat.store.SetConversation(conversation)

	return nil
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
