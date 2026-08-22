package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

var caseSummary = map[string]string{}

func askGroq(userID string, question string) (string, error) {
	if _, ok := userCase[userID]; !ok {
		userCase[userID] =
			cases[rand.Intn(len(cases))]
	}

	if _, ok := userClientType[userID]; !ok {
		userClientType[userID] =
			clientTypes[rand.Intn(len(clientTypes))]
	}

	if _, ok := userDifficulty[userID]; !ok {
		userDifficulty[userID] =
			difficulties[rand.Intn(len(difficulties))]
	}

	apiKey := os.Getenv("GROQ_API_KEY")
	fmt.Println("API KEY LENGTH:", len(apiKey))

	currentCase := userCase[userID]
	currentClientType := userClientType[userID]
	currentDifficulty := userDifficulty[userID]

	fmt.Println("Case:", currentCase)

	messages := []map[string]string{
		{
			"role": "system",
			"content": fmt.Sprintf(`
你是一位台灣法律諮詢客戶。

案件類型：
%s

客戶類型：
%s

難度：
%s

你的任務是模擬真實客戶，
讓律師透過提問逐步了解案情。

你知道完整案情，
但不會一次全部說出來。

規則：

- 你是客戶，不是律師
- 不提供法律分析或法律意見
- 只回答律師剛剛問的問題
- 不主動補充律師沒問的內容
- 不主動整理、總結或分析案件
- 不主動提供完整時間軸、證據或法律策略
- 保持案件設定一致
- 不輸出思考過程
- 不輸出<think>

客戶可能：

- 緊張
- 情緒化
- 記錯日期
- 搞不懂法律名詞
- 跳著講事情
- 回答不精確
- 只記得部分資訊

這些都是正常行為。

回答方式：

- 使用自然台灣口語
- 像 LINE 聊天
- 通常 1~5 句
- 律師問得越具體，透露越多資訊
- 不要列點
- 不要用律師或 AI 語氣

禁止使用：

本案、爭點、法律上、依據規定、綜合以上、以下幾點。

直接以客戶身份回答。
`, currentCase, currentClientType, currentDifficulty),
		},
	}
	// 加入案件摘要
	if summary, ok := caseSummary[userID]; ok {

		messages = append(
			messages,
			map[string]string{
				"role":    "system",
				"content": "目前案件摘要：\n" + summary,
			},
		)
	}

	messages = append(
		messages,
		conversations[userID]...,
	)

	messages = append(
		messages,
		map[string]string{
			"role":    "user",
			"content": question,
		},
	)

	reqBody := map[string]interface{}{
		"model":    "openai/gpt-oss-20b",
		"messages": messages,
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

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"groq status=%d body=%s",
			resp.StatusCode,
			string(body),
		)
	}

	content := result.Choices[0].Message.Content

	start := strings.Index(content, "</think>")
	if start >= 0 {
		content = content[start+8:]
	}

	content = strings.TrimSpace(content)
	conversations[userID] = append(
		conversations[userID],
		map[string]string{
			"role":    "user",
			"content": question,
		},
	)

	conversations[userID] = append(
		conversations[userID],
		map[string]string{
			"role":    "assistant",
			"content": content,
		},
	)

	if len(conversations[userID]) > 10 {
		summary, err := summarizeConversation(
			conversations[userID],
		)

		if err == nil {

			caseSummary[userID] = summary

			conversations[userID] =
				conversations[userID][len(conversations[userID])-10:]
		}
	}

	return content, nil
}

func getUserCase(userID string) string {

	currentCase, ok := userCase[userID]

	if !ok {
		currentCase = cases[rand.Intn(len(cases))]
		userCase[userID] = currentCase
	}

	return currentCase
}

func summarizeConversation(history []map[string]string) (string, error) {

	apiKey := os.Getenv("GROQ_API_KEY")

	messages := []map[string]string{
		{
			"role": "system",
			"content": `
請整理以下法律諮詢對話。

輸出格式：

【已知事實】
...

【人物】
...

【時間】
...

【金額】
...

【未確認事項】
...

限制：
- 300字內
- 不要推測
- 只整理對話中已出現資訊
`,
		},
	}

	messages = append(messages, history...)

	messages = append(messages, history...)

	reqBody := map[string]interface{}{
		"model":    "openai/gpt-oss-20b",
		"messages": messages,
	}

	jsonData, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(
		"POST",
		"https://api.groq.com/openai/v1/chat/completions",
		bytes.NewBuffer(jsonData),
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+apiKey,
	)

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(
			"groq status=%d body=%s",
			resp.StatusCode,
			string(body),
		)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
	}

	summary := strings.TrimSpace(
		result.Choices[0].Message.Content,
	)

	return summary, nil
}
