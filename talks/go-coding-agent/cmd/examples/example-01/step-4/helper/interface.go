package helper

import (
	"context"
	"go-coding-agent/pkg/client"
)

// Tool describes the features which all tools must implement.
type Tool interface {
	Call(ctx context.Context, toolCall client.ToolCall) client.D
}