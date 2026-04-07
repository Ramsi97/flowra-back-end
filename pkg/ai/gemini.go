package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiClient handles communication with Google's Gemini API.
type GeminiClient struct {
	client *genai.Client
	model  *genai.GenerativeModel
}

// NewGeminiClient initializes a new Gemini client using the provided API key.
func NewGeminiClient(ctx context.Context, apiKey string) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	model := client.GenerativeModel("gemini-1.5-flash")
	// For structured JSON output, we set the response MIME type if supported,
	// or rely on the system prompt for extraction.
	model.ResponseMIMEType = "application/json"

	return &GeminiClient{
		client: client,
		model:  model,
	}, nil
}

// Close closes the underlying connection.
func (c *GeminiClient) Close() {
	c.client.Close()
}

// AnalyzeIntent sends a prompt to Gemini and parses the JSON response into a target struct.
func (c *GeminiClient) AnalyzeIntent(ctx context.Context, systemPrompt, userPrompt string, target interface{}) error {
	c.model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text(systemPrompt)},
	}

	resp, err := c.model.GenerateContent(ctx, genai.Text(userPrompt))
	if err != nil {
		return fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return fmt.Errorf("empty response from gemini")
	}

	// Extract the text content
	part := resp.Candidates[0].Content.Parts[0]
	text, ok := part.(genai.Text)
	if !ok {
		return fmt.Errorf("unexpected response type from gemini")
	}

	rawJSON := string(text)
	// Some models might still wrap in markdown code blocks even with MIME type set
	rawJSON = strings.TrimPrefix(rawJSON, "```json")
	rawJSON = strings.TrimSuffix(rawJSON, "```")
	rawJSON = strings.TrimSpace(rawJSON)

	if err := json.Unmarshal([]byte(rawJSON), target); err != nil {
		return fmt.Errorf("failed to unmarshal gemini response: %w (raw: %s)", err, rawJSON)
	}

	return nil
}
