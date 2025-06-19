package tools

import (
	"time"

	"github.com/sashabaranov/go-openai"
)

type getTimeTool struct {
}

func (t *getTimeTool) Definition() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "get_current_time",
			Description: "Returns the current server time.",
			Parameters: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
	}
}
func (t *getTimeTool) Handle(call openai.ToolCall) (string, error) {
	// fmt.Println("getTimeTool Handle called with call:", call)
	if call.Function.Name != "get_current_time" {
		return "", nil // or an error if you want to handle unknown calls
	}
	// Call the function to get the current time
	result := time.Now().Format(time.RFC1123)

	// Return the result in the expected format
	return result, nil
}
