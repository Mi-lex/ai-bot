package chat

type Context struct {
	// user, assistant, system
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Conversation struct {
	Id          string  `json:"id"`
	Model       string  `json:"model"`
	Temperature float32 `json:"temperature"`

	ContextList []Context `json:"contextList"`
}

func NewConversation(id string, model string, temperature float32) *Conversation {
	return &Conversation{
		Id:          id,
		Model:       model,
		Temperature: temperature,
	}
}

func (conversation *Conversation) addContent(role string, content string) {
	conversation.ContextList = append(conversation.ContextList, Context{
		Role:    role,
		Content: content,
	})
}

func (conversation *Conversation) AddUserContext(content string) {
	conversation.addContent("user", content)
}

func (conversation *Conversation) AddAssistantContent(content string) {
	conversation.addContent("assistant", content)
}

func (conversation *Conversation) AddSystemContent(content string) {
	conversation.addContent("system", content)
}
