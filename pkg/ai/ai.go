package ai

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/ferama/yay/pkg/ai/tools"
	"github.com/sashabaranov/go-openai"
)

const (
	// input must not exceed this size
	maxCharsGPT = 12250

	// if a request fail, how many times we should retry
	maxRetries = 3

	prompt = `Tou are an assistant that thinks step-by-step and uses tools to answer questions.

When you get a question:
1. Think about what to do using <think>.
2. If you need a tool, call it using <tool>tool_name: input</tool>.
3. Always show the tool calls and final answer in the output.
4. End with <final>answer</final>

Example:
<question>What is 2 + 2?</question>
<think>I should use a calculator.</think>
<final>
The answer is 4
</final>

<question>What's the latest news about LangChain?</question>
<think>I should search the internet for this.</think>
<final>
[insert result here]
</final>

<question>Tell me about LangChain from Wikipedia and summarize it.</question>
<think>I should use Wikipedia search and then summarize the info.</think>
<final>
[insert summarized info]
</final>

Now answer this:
`
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

	client := openai.NewClient(os.Getenv("YAY_API_KEY"))

	// allow usage of local opeani comptatible servers like
	// python-llama-cpp or vllm
	// Example:
	//	YAY_API_BASEURL="http://localhost:8000/v1"

	customUrl := os.Getenv("YAY_API_BASEURL")
	if customUrl != "" {
		config := openai.DefaultConfig(os.Getenv("YAY_API_KEY"))
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
	content = fmt.Sprintf("%s\n<question>%s</question>", prompt, content)

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

	msg := ""
	if len(resp.Choices[0].Message.ToolCalls) > 0 {
		msg, err = a.handleTools(resp.Choices[0].Message.ToolCalls, resp.Choices[0].Message)
		if err != nil {
			return "", fmt.Errorf("tool handling error: %v", err)
		}
	} else {
		msg = resp.Choices[0].Message.Content
		a.messages = append(a.messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: msg,
		})
	}

	msg = strings.ReplaceAll(msg, "<think>", "# Thinking\n")
	msg = strings.ReplaceAll(msg, "</think>", "")
	msg = strings.ReplaceAll(msg, "<final>", "# Answer\n")
	msg = strings.ReplaceAll(msg, "</final>", "")

	return msg, nil
}
