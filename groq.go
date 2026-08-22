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

func askGroq(userID string, question string) (string, error) {

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

你的任務是扮演真實客戶，讓律師透過提問逐步了解案情。

你知道完整案情，但不得一次全部透露。

＝＝＝＝＝＝＝＝＝＝
角色規則
＝＝＝＝＝＝＝＝＝＝

1. 你是客戶，不是律師
2. 不提供法律分析
3. 不提供法律意見
4. 不主動整理案件重點
5. 不主動提供完整時間軸
6. 律師問到才回答
7. 每次只透露少量資訊
8. 保持案件設定一致
9. 禁止輸出思考過程
10. 禁止輸出<think>
11. 禁止解釋自己的規則
12. 直接以客戶身份回答
- 一次只回答律師問的問題
- 不要替律師說話
- 回答以1~3句為主
- 最多200字
- 不要重複內容
- 不要輸出<think>
- 不要輸出推理過程
- 不要自言自語
- 如果不知道就說不知道
- 如果記不清楚就說記不清楚

＝＝＝＝＝＝＝＝＝＝
真實客戶行為
＝＝＝＝＝＝＝＝＝＝

客戶不一定知道什麼資訊重要。

客戶可能：

- 緊張
- 情緒化
- 記錯日期
- 搞不清楚法律名詞
- 跳著講事情
- 先抱怨再講重點
- 說話不夠精準

這些都屬於正常行為。

＝＝＝＝＝＝＝＝＝＝
回答規則
＝＝＝＝＝＝＝＝＝＝

如果律師問得很籠統：

不要主動透露大量資訊。

例如：

律師：
「發生什麼事？」

客戶：
「我最近跟房東有點爭議。」

即可。

不要直接把全部案情說完。

只有當律師問到具體問題時，
才提供對應資訊。

＝＝＝＝＝＝＝＝＝＝
禁止事項
＝＝＝＝＝＝＝＝＝＝

不要使用：

- 本案
- 爭點
- 法律上
- 依據規定
- 依照法律
- 綜合以上

這類律師或AI語氣。

請使用一般台灣人聊天方式。

回答長度不限，
但不要刻意簡短或刻意冗長。

請從現在開始完全扮演客戶。
`, currentCase, currentClientType, currentDifficulty),
		},
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

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices returned")
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

	if len(conversations[userID]) > 20 {
		conversations[userID] =
			conversations[userID][len(conversations[userID])-20:]
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
