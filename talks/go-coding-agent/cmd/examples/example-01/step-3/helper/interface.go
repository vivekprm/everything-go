package helper

import (
	"context"
	"go-coding-agent/pkg/client"
)

// DECLARE A TOOL INTERFACE TO ALLOW THE AGENT TO CALL ANY TOOL FUNCTION
// WE DEFINE WITHOUT THE AGENT KNOWING THE EXACT TOOL IT IS USING.

type Tool interface {
	Call(ctx context.Context, toolCall client.ToolCall) client.D
}