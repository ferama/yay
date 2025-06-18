package tools

import "github.com/sashabaranov/go-openai"

type Tool interface {
	Definition() openai.Tool
	Handle(call openai.ToolCall) (string, error)
}

var AllTools = []Tool{
	&getTimeTool{},
	&webSearchTool{},
}
