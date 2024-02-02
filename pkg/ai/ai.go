package ai

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/sashabaranov/go-openai"
)

const (
	// is the model that we are actually using
	openAIModel = openai.GPT3Dot5Turbo

	// input must not exceed this size
	maxCharsGPT = 12250

	// if a request fail, how many times we should retry
	maxRetries = 3

	formatHeader = "format the response as markdown."
)

// define errors
var (
	ErrInvalidApiKey = errors.New("invalid api key")
	ErrRateLimit     = errors.New("rate limit error")
	ErrMaxPromptSize = errors.New("maximum prompt size exceeded")
	ErrModelNotFound = errors.New("model not found")
)

type AI struct {
	messages []openai.ChatCompletionMessage
	client   *openai.Client
}

func NewAI() *AI {

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	// allow usage of local opeani comptatible servers like
	// python-llama-cpp or vllm
	// Example:
	//	YAY_API_BASEURL="http://localhost:8000/v1"

	customUrl := os.Getenv("YAY_API_BASEURL")
	if customUrl != "" {
		client = openai.NewClientWithConfig(openai.ClientConfig{
			BaseURL:    customUrl,
			HTTPClient: &http.Client{},
		})
	}

	ai := &AI{
		messages: make([]openai.ChatCompletionMessage, 0),
		client:   client,
	}
	return ai
}

func (a *AI) SendMsg(content string) (string, error) {
	content = fmt.Sprintf("%s %s", formatHeader, content)

	if len(content) > maxCharsGPT {
		// limit input size
		content = content[:maxCharsGPT]
	}

	a.messages = append(a.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: content,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// this actually makes the call to the api
	doreq := func() (openai.ChatCompletionResponse, error) {
		return a.client.CreateChatCompletion(
			ctx,
			openai.ChatCompletionRequest{
				Model:    openAIModel,
				Messages: a.messages,
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
			return "", ErrModelNotFound
		case http.StatusBadRequest:
			if ae.Code == "context_length_exceeded" {
				return "", ErrMaxPromptSize
			}
		case http.StatusUnauthorized:
			return "", ErrInvalidApiKey
		case http.StatusTooManyRequests:
			return "", ErrRateLimit
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
