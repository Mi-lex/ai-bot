package chat

import (
	"encoding/json"
	"fmt"

	"github.com/Mi-lex/dgpt-bot/utils"
)

const conversationKeyPrefix = "conversation"
const modelKey = "modelName"

func getStoreKey(id string) string {
	return conversationKeyPrefix + ":" + id
}

type Store struct {
	redis *utils.Redis
}

func (s *Store) SetModel(model string) error {
	err := s.redis.Set(modelKey, model)

	if err != nil {
		return fmt.Errorf("Failed to set model %s, %w", model, err)
	}

	return nil
}

func (s *Store) GetModel() (string, error) {
	result, err := s.redis.Get(modelKey)

	if err != nil {
		return "", err
	}

	return result, nil
}

func (s *Store) GetConversation(id string) (conversation *Conversation, err error) {
	conversationJson, err := s.redis.GetJson(getStoreKey(id))

	if err != nil {
		// there is no record with given key

		return nil, nil
	}

	err = json.Unmarshal(conversationJson.([]byte), &conversation)

	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal stored conversation: %w", err)
	}

	return conversation, nil
}

func (s *Store) SetConversation(conversation *Conversation) error {
	_, err := s.redis.SetJson(getStoreKey(conversation.Id), conversation)

	if err != nil {
		return fmt.Errorf("Failed to set stored conversation %s: %w", conversation.Id, err)
	}

	return nil
}
