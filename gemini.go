package main

import (
	"context"

	"google.golang.org/genai"
)

const (
	modelPro   = "gemini-3-pro-preview"
	modelFlash = "gemini-3-flash-preview"

	defaultGeminiOPRef = "op://Private/Gemini API Key/credential"
)

type geminiProvider struct {
	client *genai.Client
}

func newGeminiProvider(ctx context.Context) (*geminiProvider, error) {
	apiKey, err := resolveKey("GEMINI_API_KEY", defaultGeminiOPRef)
	if err != nil {
		return nil, err
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	return &geminiProvider{client: client}, nil
}

func (g *geminiProvider) Classify(ctx context.Context, text string) (string, error) {
	resp, err := g.client.Models.GenerateContent(ctx, modelFlash,
		genai.Text(text),
		&genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{{Text: classifyPrompt}},
			},
			MaxOutputTokens: genai.Ptr[int32](16),
		},
	)
	if err != nil {
		return "", err
	}

	return parseClassification(resp.Text()), nil
}

func (g *geminiProvider) Rewrite(ctx context.Context, model, systemPrompt, text string) (string, error) {
	resp, err := g.client.Models.GenerateContent(ctx, model,
		genai.Text(text),
		&genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{{Text: systemPrompt}},
			},
			MaxOutputTokens: genai.Ptr(maxOutputTokens),
		},
	)
	if err != nil {
		return "", err
	}

	return resp.Text(), nil
}

func (g *geminiProvider) Prompt(ctx context.Context, model, text string) (string, error) {
	resp, err := g.client.Models.GenerateContent(ctx, model,
		genai.Text(text),
		&genai.GenerateContentConfig{
			MaxOutputTokens: genai.Ptr(maxOutputTokens),
		},
	)
	if err != nil {
		return "", err
	}

	return resp.Text(), nil
}

func (g *geminiProvider) DefaultModel(contentType string) string {
	switch contentType {
	case "code", "config", "prompt":
		return modelPro
	default:
		return modelFlash
	}
}
