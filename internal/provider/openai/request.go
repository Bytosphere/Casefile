package openai

type MessageRequest struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string           `json:"model"`
	Messages []MessageRequest `json:"messages"`
}

func NewChatRequest(model, prompt string) ChatRequest {
	return ChatRequest{
		Model: model,
		Messages: []MessageRequest{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}
}
