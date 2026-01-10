package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// Client wraps the LLM client
type Client struct {
	llm   llms.Model
	model string // Model name for API calls
}

// NewClient creates a new LLM client
func NewClient(provider, apiKey, url, modelName string) (*Client, error) {
	var llmModel llms.Model
	var err error

	switch provider {
	case "openai":
		if apiKey == "" {
			apiKey = os.Getenv("OPENAI_API_KEY")
		}
		opts := []openai.Option{
			openai.WithToken(apiKey),
		}
		// Add custom URL if provided
		if url != "" {
			opts = append(opts, openai.WithBaseURL(url))
		}
		llmModel, err = openai.New(opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create OpenAI client: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", provider)
	}

	return &Client{llm: llmModel, model: modelName}, nil
}

// Generate generates text from a prompt
func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	var options []llms.CallOption
	// Add model name if configured
	if c.model != "" {
		options = append(options, llms.WithModel(c.model))
	}
	completion, err := c.llm.Call(ctx, prompt, options...)
	if err != nil {
		return "", fmt.Errorf("LLM generation failed: %w", err)
	}
	return completion, nil
}

// GenerateWithTools generates text with tool calling support
func (c *Client) GenerateWithTools(ctx context.Context, prompt string, tools []llms.Tool) (string, []llms.ToolCall, error) {
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, prompt),
	}

	options := []llms.CallOption{
		llms.WithTools(tools),
	}

	// Add model name if configured
	if c.model != "" {
		options = append(options, llms.WithModel(c.model))
	}

	response, err := c.llm.GenerateContent(ctx, messages, options...)
	if err != nil {
		return "", nil, fmt.Errorf("LLM generation with tools failed: %w", err)
	}

	// Extract text response
	var textResponse string
	if len(response.Choices) > 0 {
		textResponse = response.Choices[0].Content
	}

	// Extract tool calls from response
	var toolCalls []llms.ToolCall
	if len(response.Choices) > 0 {
		toolCalls = response.Choices[0].ToolCalls
	}

	return textResponse, toolCalls, nil
}

// GetModel returns the underlying LLM model
func (c *Client) GetModel() llms.Model {
	return c.llm
}
