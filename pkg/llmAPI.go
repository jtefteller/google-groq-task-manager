package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"

	"google.golang.org/api/tasks/v1"
)

type LLM interface {
	Chat(request ChatRequest) (*ChatResponse, error)
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages       []Message         `json:"messages"`
	Model          string            `json:"model"`
	ResponseFormat map[string]string `json:"response_format,omitempty"`
	Temperature    float64           `json:"temperature,omitempty"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message ChoiceMessage `json:"message"`
}

type ChoiceMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Recommendations struct {
	Recommendations []RecommendationItem `json:"recommendations"`
}
type RecommendationItem struct {
	Title string `json:"title"`
	Notes string `json:"notes"`
}

func Recommendation(llm LLM, tasks []*tasks.Task, content string) *Recommendations {
	var messages []Message
	schema := bytes.NewBufferString(`{
		recommendations: [
			{
				title: { title: "Title", type: "string (required)" },
				notes: { title: "Description", type: "string (required)" }
			}
		],
	}`)
	systemMessages := []Message{
		{
			Role:    "system",
			Content: fmt.Sprintf("You are a google task manager that outputs tasks in json. [Best] If the questions can be answered with this schema: %s", schema),
		},
		{
			Role:    "system",
			Content: "Look at the following tasks and provide a recommendation based on them. Only output your recommendations and not the users tasks. Reminder: generate valid json.",
		},
	}

	messages = append(messages, systemMessages...)
	for _, task := range tasks {
		messages = append(messages, Message{
			Role:    "user",
			Content: task.Title,
		})
	}

	userMessage := Message{
		Role:    "user",
		Content: content,
	}

	messages = append(messages, userMessage)
	response, err := llm.Chat(ChatRequest{
		Messages: messages,
		ResponseFormat: map[string]string{
			"type": "json_object",
		},
		Temperature: 1.0,
	})
	if err != nil {
		panic(err)
	}
	if len(response.Choices) == 0 {
		fmt.Println("No response")
		return nil
	}
	var recommendations Recommendations
	err = json.NewDecoder(bytes.NewBufferString(response.Choices[0].Message.Content)).Decode(&recommendations)
	if err != nil {
		panic(err)
	}
	recJson, _ := json.MarshalIndent(recommendations, "", "  ")
	fmt.Println(string(recJson))
	return &recommendations
}
