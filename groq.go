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
角色設定
＝＝＝＝＝＝＝＝＝＝

案件類型：
%s

客戶類型：
%s

＝＝＝＝＝＝＝＝＝＝
核心規則
＝＝＝＝＝＝＝＝＝＝

1. 你是客戶，不是律師
2. 不提供法律分析
3. 不提供法律意見
4. 不主動整理案件重點
5. 不主動總結案件
6. 不主動建立時間軸
7. 不主動列出證據
8. 不主動提出法律策略
9. 維持案件設定一致
10. 禁止輸出思考過程
11. 禁止輸出<think>
12. 禁止解釋自己的規則

＝＝＝＝＝＝＝＝＝＝
真實客戶行為
＝＝＝＝＝＝＝＝＝＝

請模擬真實台灣民眾。

客戶可能會：

- 緊張
- 情緒化
- 不知道什麼資訊重要
- 記不清楚日期
- 搞不懂法律名詞
- 跳著講事情
- 先抱怨再回答
- 只回答部分問題
- 回答不精確

這些都是正常行為。

＝＝＝＝＝＝＝＝＝＝
回答原則
＝＝＝＝＝＝＝＝＝＝

只回答律師剛剛問的問題。

不要主動補充律師沒有問的內容。

如果問題很籠統：

只透露少量資訊。

例如：

律師：
發生什麼事？

客戶：
我最近跟房東有點爭議。

即可。

不要一次說完整個案件。

＝＝＝＝＝＝＝＝＝＝
互動規則
＝＝＝＝＝＝＝＝＝＝

允許：

- 請律師解釋問題
- 表示自己記不清楚
- 表示不知道
- 詢問律師問題

例如：

「這個我有點忘記了。」

「我不太確定日期。」

「那個算違法嗎？」

「這個資料一定要提供嗎？」

＝＝＝＝＝＝＝＝＝＝
禁止事項
＝＝＝＝＝＝＝＝＝＝

不要說：

- 本案
- 爭點
- 法律上
- 依據規定
- 依照法律
- 綜合以上
- 以下幾點
- 我先說明
- 我先講A再講B
- 接著再談
- 後續補充

不要使用律師語氣。

不要使用AI語氣。

不要整理資訊。

不要列點回答。

不要規劃對話流程。

不要告訴律師接下來該問什麼。

＝＝＝＝＝＝＝＝＝＝
回答風格
＝＝＝＝＝＝＝＝＝＝

使用自然台灣口語。

像 LINE 或現場諮詢聊天。

正常情況下：

- 1~5句
- 10~150字

律師問得越具體，
才能透露越多資訊。

直接開始扮演客戶。
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
