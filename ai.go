package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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

	maxRetries := 3
	doreq := func() (openai.ChatCompletionResponse, error) {
		return a.client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    openai.GPT3Dot5Turbo,
				Messages: a.messages,
			},
		)
	}
	retry := func() (openai.ChatCompletionResponse, error) {
		for {
			r, err := doreq()
			if err == nil {
				return r, err
			}
			maxRetries--
			if maxRetries <= 0 {
				break
			}
		}
		return openai.ChatCompletionResponse{}, fmt.Errorf("no more retries")
	}

	resp, err := doreq()

	ae := &openai.APIError{}
	if errors.As(err, &ae) {
		switch ae.HTTPStatusCode {
		case http.StatusNotFound:
			return "", fmt.Errorf("openai server error: model not found")
		case http.StatusBadRequest:
			if ae.Code == "context_length_exceeded" {
				return "", fmt.Errorf("maximum prompt size exceeded")
			}
		case http.StatusUnauthorized:
			return "", fmt.Errorf("invalid api key")
		case http.StatusTooManyRequests:
			return "", fmt.Errorf("rate limit error")
		case http.StatusInternalServerError:
			resp, err = retry()
		default:
			resp, err = retry()
		}
	}

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
