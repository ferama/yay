package ai

import (
	"errors"
	"fmt"
	"os"

	"github.com/ferama/yay/pkg/ai/tools"
	"github.com/sashabaranov/go-openai"
)

const (
	// is the model that we are actually using
	// openAIModel = openai.GPT3Dot5Turbo
	openAIModel = "meta-llama/llama-4-maverick-17b-128e-instruct"
	// openAIModel = "llama-3.3-70b-versatile"

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
		config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
		config.BaseURL = customUrl
		client = openai.NewClientWithConfig(config)
	}

	ai := &AI{
		messages: make([]openai.ChatCompletionMessage, 0),
		client:   client,
	}
	return ai
}

func (a *AI) handleToolResponse(message openai.ChatCompletionMessage, result string, call openai.ToolCall) (string, error) {
	a.messages = append(a.messages,
		message, // tool call message from assistant
		openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			Name:       call.Function.Name,
			Content:    result,
			ToolCallID: call.ID,
		},
	)

	finalRes, err := doRequest(a.client, a.messages, false)
	if err != nil {
		return "", fmt.Errorf("chat error: %v", err)
	}

	msg := finalRes.Choices[0].Message.Content
	a.messages = append(a.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: msg,
	})
	return msg, nil
}

func (a *AI) handleTools(calls []openai.ToolCall, message openai.ChatCompletionMessage) (string, error) {
	for _, call := range calls {
		for _, tool := range tools.AllTools {
			res, err := tool.Handle(call) // call the tool handler
			if err != nil {
				return "", fmt.Errorf("tool call error: %v", err)
			}
			if res != "" {
				// if the tool returns a result, handle it
				return a.handleToolResponse(message, res, call)
			}
		}
	}
	return "", fmt.Errorf("no tool calls found")
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

	resp, err := doRequest(a.client, a.messages, true)

	if err != nil {
		return "", fmt.Errorf("chat error: %v", err)
	}

	if len(resp.Choices[0].Message.ToolCalls) > 0 {
		return a.handleTools(resp.Choices[0].Message.ToolCalls, resp.Choices[0].Message)
	}

	msg := resp.Choices[0].Message.Content
	a.messages = append(a.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: msg,
	})
	return msg, nil
}
