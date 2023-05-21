package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

type AI struct {
	messages []openai.ChatCompletionMessage
	client   *openai.Client
}

func newAI() *AI {
	ai := &AI{
		messages: make([]openai.ChatCompletionMessage, 0),
		client:   openai.NewClient(os.Getenv("OPENAI_API_KEY")),
	}
	return ai
}

func (a *AI) sendMsg(content string) (string, error) {
	a.messages = append(a.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: content,
	})

	resp, err := a.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT3Dot5Turbo,
			Messages: a.messages,
		},
	)

	if err != nil {
		return "", fmt.Errorf("chat error: %v", err)
	}

	msg := resp.Choices[0].Message.Content
	a.messages = append(a.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: msg,
	})
	return msg, nil
}
