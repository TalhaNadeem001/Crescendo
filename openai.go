// openai.go - Calls OpenAI API to break a task into 3 simpler subtasks.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// openaiRequest and openaiResponse match the Chat Completions API.
type openaiRequest struct {
	Model    string          `json:"model"`
	Messages []openaiMessage `json:"messages"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// BreakIntoSubtasks calls the OpenAI API to break the given task into exactly 3 simpler subtasks.
// Returns up to 3 non-empty trimmed lines from the model response, or an error.
func BreakIntoSubtasks(task string, apiKey string) ([]string, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_KEY is not set")
	}

	prompt := fmt.Sprintf(`Break down the following task into exactly 3 simpler subtasks. Return only the 3 subtasks, one per line. No numbering, bullets, or extra text.

Task: %s`, task)

	reqBody := openaiRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openaiMessage{
			{Role: "user", Content: prompt},
		},
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai api error %d: %s", resp.StatusCode, string(respBytes))
	}

	var apiResp openaiResponse
	if err := json.Unmarshal(respBytes, &apiResp); err != nil {
		return nil, err
	}
	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("openai returned no choices")
	}

	content := apiResp.Choices[0].Message.Content
	var out []string
	for _, line := range strings.Split(content, "\n") {
		s := strings.TrimSpace(line)
		// Strip leading number/bullet like "1." or "-"
		if idx := strings.IndexAny(s, ".-)"); idx == 0 || (idx == 1 && s[0] >= '0' && s[0] <= '9') {
			s = strings.TrimSpace(s[idx+1:])
		}
		if s != "" {
			out = append(out, s)
			if len(out) >= 3 {
				break
			}
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("could not parse subtasks from response")
	}
	return out, nil
}
