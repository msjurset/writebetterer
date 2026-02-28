package main

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

const (
	claudeModelSonnet = "claude-sonnet-4-20250514"
	claudeModelHaiku  = "claude-haiku-4-5-20251001"

	defaultClaudeOPRef = "op://Private/Anthropic API Key/credential"
)

type claudeProvider struct {
	client anthropic.Client
}

func newClaudeProvider() (*claudeProvider, error) {
	apiKey, err := resolveKey("ANTHROPIC_API_KEY", defaultClaudeOPRef)
	if err != nil {
		return nil, err
	}

	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	return &claudeProvider{client: client}, nil
}

func (c *claudeProvider) Classify(ctx context.Context, text string) (string, error) {
	msg, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(claudeModelHaiku),
		MaxTokens: 16,
		System: []anthropic.TextBlockParam{
			{Text: classifyPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(text)),
		},
	})
	if err != nil {
		return "", err
	}

	return parseClassification(messageText(msg)), nil
}

func (c *claudeProvider) Rewrite(ctx context.Context, model, systemPrompt, text string) (string, error) {
	msg, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: int64(maxOutputTokens),
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(text)),
		},
	})
	if err != nil {
		return "", err
	}

	return messageText(msg), nil
}

func (c *claudeProvider) Prompt(ctx context.Context, model, text string) (string, error) {
	msg, err := c.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: int64(maxOutputTokens),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(text)),
		},
	})
	if err != nil {
		return "", err
	}

	return messageText(msg), nil
}

func (c *claudeProvider) DefaultModel(contentType string) string {
	switch contentType {
	case "code", "config", "prompt":
		return claudeModelSonnet
	default:
		return claudeModelHaiku
	}
}

// messageText extracts the concatenated text from a Claude message response.
func messageText(msg *anthropic.Message) string {
	var result string
	for _, block := range msg.Content {
		if block.Type == "text" {
			result += block.Text
		}
	}
	return result
}

// Verify interface compliance at compile time.
var _ Provider = (*claudeProvider)(nil)
var _ Provider = (*geminiProvider)(nil)

