package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func askGroq(question string) (string, error) {

	apiKey := os.Getenv("GROQ_API_KEY")
	fmt.Println("API KEY LENGTH:", len(apiKey))
	reqBody := map[string]interface{}{
		"model": "qwen/qwen3.6-27b",
		"messages": []map[string]string{
			{
				"role": "system",
				"content": `
你是一位台灣法律諮詢客戶。

規則：

1. 扮演客戶，不是律師
2. 不提供法律意見
3. 不要一次透露全部案情
4. 像真人聊天
5. 回答控制在100字內
6. 律師問到才補充細節
7. 維持角色
8. 禁止輸出思考過程
9. 禁止輸出<think>
10. 直接回答
`,
			},
			{
				"role":    "user",
				"content": question,
			},
		},
	}

	jsonData, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(
		"POST",
		"https://api.groq.com/openai/v1/chat/completions",
		bytes.NewBuffer(jsonData),
	)

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	fmt.Println("Groq Status:", resp.StatusCode)
	fmt.Println("Groq Response:", string(body))

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	json.Unmarshal(body, &result)

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}

	return result.Choices[0].Message.Content, nil
}
