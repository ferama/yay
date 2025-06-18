package ai

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/ferama/yay/pkg/ai/tools"
	"github.com/sashabaranov/go-openai"
)

func doRequest(client *openai.Client, messages []openai.ChatCompletionMessage, useTools bool) (*openai.ChatCompletionResponse, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	AIModel, exists := os.LookupEnv("YAY_MODEL")
	if !exists {
		AIModel = openai.GPT3Dot5Turbo
	}

	t := []openai.Tool{}
	if useTools {
		// t = []openai.Tool{tools.GetTimeTool}
		for _, tool := range tools.AllTools {
			t = append(t, tool.Definition())
		}
	}
	// this actually makes the call to the api
	doreq := func() (openai.ChatCompletionResponse, error) {
		return client.CreateChatCompletion(
			ctx,
			openai.ChatCompletionRequest{
				Model:    AIModel,
				Messages: messages,
				Tools:    t,
			},
		)
	}

	// retries is something went wrong
	retries := maxRetries
	retry := func() (openai.ChatCompletionResponse, error) {
		for {
			r, err := doreq()
			if err == nil {
				return r, err
			}
			retries--
			if retries <= 0 {
				break
			}
		}
		return openai.ChatCompletionResponse{}, fmt.Errorf("no more retries")
	}

	resp, err := doreq()

	// manage errors
	ae := &openai.APIError{}
	if errors.As(err, &ae) {
		switch ae.HTTPStatusCode {
		case http.StatusNotFound:
			return nil, ErrModelNotFound
		case http.StatusBadRequest:
			if ae.Code == "context_length_exceeded" {
				return nil, ErrMaxPromptSize
			}
		case http.StatusUnauthorized:
			return nil, ErrInvalidApiKey
		case http.StatusTooManyRequests:
			return nil, ErrRateLimit
		case http.StatusInternalServerError:
			resp, err = retry()
		default:
			resp, err = retry()
		}
	}

	if err != nil {
		return nil, fmt.Errorf("chat error: %v", err)
	}

	return &resp, nil
}
