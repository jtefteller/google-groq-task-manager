package pkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

const (
	CHAT_URL = "https://api.groq.com/openai/v1/chat/completions"
)

type Groq struct {
	APIKey string
	Model  string
}

func NewGroq(apiKey string) *Groq {
	return &Groq{
		APIKey: apiKey,
		Model:  "llama3-70b-8192",
		// Model:  "llama-3.1-70b-versatile",
	}
}

func (g *Groq) Chat(request ChatRequest) (*ChatResponse, error) {
	request.Model = g.Model

	httpClient := &http.Client{}

	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	bodyReader := bytes.NewReader(body)

	req, err := http.NewRequest("POST", CHAT_URL, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.APIKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Print(resp.Status)
		b := [1000]byte{}
		n, err := resp.Body.Read(b[:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Print(string(b[:n]))
				return nil, err
			}
			log.Fatal(err)
		}
		log.Println(string(b[:n]))
	}

	var chatResponse ChatResponse

	err = json.NewDecoder(resp.Body).Decode(&chatResponse)
	if err != nil {
		return nil, err
	}

	return &chatResponse, nil
}
