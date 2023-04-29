package chatgpt

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"time"

	jsoniter "github.com/json-iterator/go"
)

type OfficialChatGPT struct {
	system_prompt string
	token         string
	http          http.Client
}

func NewOfficialChatGPTClient(timeout time.Duration) ChatGPTClient {
	return &OfficialChatGPT{
		token: os.Getenv("CHATGPT_TOKEN"),
		http:  http.Client{Timeout: timeout},
	}
}

func (c *OfficialChatGPT) WithSystemPrompt(system_prompt string) {
	c.system_prompt = system_prompt
}

func (c *OfficialChatGPT) ChatCompletion(prompt string) (string, error) {
	messages := []Message{
		{
			Role:    "user",
			Content: prompt,
		},
	}

	if c.system_prompt != "" {
		messages = append([]Message{
			{
				Role:    "system",
				Content: c.system_prompt,
			},
		}, messages...)
	}

	body, err := jsoniter.Marshal(map[string]interface{}{
		"model":    os.Getenv("CHATGPT_MODEL"),
		"messages": messages,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", os.Getenv("CHATGPT_CHAT_COMPLETION_URL"), bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		var err map[string]interface{}
		jsoniter.NewDecoder(resp.Body).Decode(&err)
		return "", fmt.Errorf("[%d] %v", resp.StatusCode, err)
	}

	var data ResponseData
	err = jsoniter.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}

	if len(data.Choices) == 0 {
		return "", fmt.Errorf("no choices")
	}

	fmt.Printf("RESPONSE FROM CHATGPT %v\n\n\n", data.Choices[0].Message)

	return data.Choices[0].Message.Content, nil
}
