package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type OpenAIAssistant struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
	noop    *NoopAssistant
}

func NewOpenAIAssistant(baseURL string, apiKey string, model string, client *http.Client) *OpenAIAssistant {
	if client == nil {
		client = http.DefaultClient
	}
	return &OpenAIAssistant{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		model:   model,
		client:  client,
		noop:    NewNoopAssistant(),
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

func (a *OpenAIAssistant) complete(ctx context.Context, system string, user string) (string, error) {
	payload, err := json.Marshal(chatRequest{
		Model: a.model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("ai provider returned status %d", resp.StatusCode)
	}

	var decoded chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", err
	}
	if len(decoded.Choices) == 0 {
		return "", fmt.Errorf("ai provider returned no choices")
	}

	return strings.TrimSpace(decoded.Choices[0].Message.Content), nil
}

func (a *OpenAIAssistant) Rewrite(ctx context.Context, text string) (string, error) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", ErrEmptyText
	}
	return a.complete(ctx,
		"Rewrite the user's text into grammatically correct, professionally phrased prose. Preserve the meaning. Do not add new claims. Reply with the rewritten text only.",
		trimmed,
	)
}

func (a *OpenAIAssistant) Chat(ctx context.Context, section Section, message string) (string, error) {
	system := fmt.Sprintf(
		"You are the GamiDoc assistant. The user is filling in the '%s' section of a gamification design document. Give concise, concrete guidance for this section. Clarify domain terminology when asked.",
		section.Name,
	)
	return a.complete(ctx, system, message)
}

func (a *OpenAIAssistant) RecommendBranch(ctx context.Context, spark string) (string, error) {
	answer, err := a.complete(ctx,
		"The user wrote a free-form project idea. Answer with the single letter A if the text mostly concerns user experience, personas, or journeys; answer B if it mostly concerns game mechanics, rewards, or technology. Answer with A or B only.",
		spark,
	)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(strings.ToUpper(answer), "B") {
		return "B", nil
	}
	return "A", nil
}

func (a *OpenAIAssistant) Prefill(ctx context.Context, spark string, sections []Section) (map[int]json.RawMessage, error) {
	trimmed := strings.TrimSpace(spark)
	if trimmed == "" {
		return map[int]json.RawMessage{}, nil
	}

	var names []string
	for _, section := range sections {
		names = append(names, fmt.Sprintf("%d: %s", section.Number, section.Name))
	}
	system := fmt.Sprintf(
		"From the user's project idea, draft initial content for the sections of a gamification design document: %s. Reply with a JSON object whose keys are the section numbers as strings and whose values are objects with a single 'draft' string field. Include only sections the idea gives real material for. Reply with JSON only.",
		strings.Join(names, "; "),
	)

	answer, err := a.complete(ctx, system, trimmed)
	if err != nil {
		return nil, err
	}
	answer = strings.TrimPrefix(answer, "```json")
	answer = strings.Trim(answer, "` \n")

	var decoded map[string]json.RawMessage
	if err := json.Unmarshal([]byte(answer), &decoded); err != nil {
		return a.noop.Prefill(ctx, spark, sections)
	}

	result := make(map[int]json.RawMessage, len(decoded))
	for key, value := range decoded {
		number, err := strconv.Atoi(key)
		if err != nil {
			continue
		}
		result[number] = value
	}
	return result, nil
}

func (a *OpenAIAssistant) Enhance(ctx context.Context, sections []SectionText) ([]SectionText, error) {
	result := make([]SectionText, 0, len(sections))
	for _, section := range sections {
		if strings.TrimSpace(section.Text) == "" {
			continue
		}
		enhanced, err := a.complete(ctx,
			fmt.Sprintf(
				"Rewrite the raw notes for the '%s' section of a gamification design report into naturally readable prose. Preserve the meaning and facts. Do not add new claims. Remove duplicated statements. Reply with the prose only.",
				section.Name,
			),
			section.Text,
		)
		if err != nil {
			return nil, err
		}
		section.Text = enhanced
		result = append(result, section)
	}
	return result, nil
}
