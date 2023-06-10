package chat

type Context struct {
	// user, assistant, system
	role    string
	content string
}

type Conversation struct {
	Id          string
	Model       string
	Temperature float32

	contextList []Context
}

func NewConversation(id string, model string, temperature float32) *Conversation {
	return &Conversation{
		Id:          id,
		Model:       model,
		Temperature: temperature,
	}
}

func (conversation *Conversation) addContent(role string, content string) {
	conversation.contextList = append(conversation.contextList, Context{
		role:    role,
		content: content,
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
