package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
)

type WebhookRequest struct {
	Events []struct {
		ReplyToken string `json:"replyToken"`

		Source struct {
			UserID string `json:"userId"`
		} `json:"source"`

		Message struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"message"`
	} `json:"events"`
}

var conversations = map[string][]map[string]string{}
var cases = []string{
	"車禍糾紛",
	"勞資糾紛",
	"離婚案件",
	"監護權爭議",
	"租賃糾紛",
	"借貸糾紛",
	"詐騙案件",
	"網購糾紛",
	"侵權損害賠償",
	"遺產繼承",
}
var clientTypes = []string{
	"理性型",
	"情緒型",
	"健談型",
	"寡言型",
	"強勢型",
	"緊張型",
	"不信任型",
	"記憶模糊型",
	"愛抱怨型",
	"自以為懂法律型",
}
var difficulties = []string{
	"簡單",
	"普通",
	"困難",
}
var userCase = map[string]string{}
var userClientType = map[string]string{}
var userDifficulty = map[string]string{}

func webhook(w http.ResponseWriter, r *http.Request) {

	body, _ := io.ReadAll(r.Body)

	fmt.Println("=== WEBHOOK HIT ===")
	fmt.Println(string(body))

	var req WebhookRequest
	json.Unmarshal(body, &req)

	if len(req.Events) > 0 {

		userText := req.Events[0].Message.Text
		userID := req.Events[0].Source.UserID

		fmt.Println("user:", userText)
		fmt.Println("userID:", userID)

		if userText == "/new" {

			delete(conversations, userID)

			caseType := cases[rand.Intn(len(cases))]
			clientType := clientTypes[rand.Intn(len(clientTypes))]
			difficulty := difficulties[rand.Intn(len(difficulties))]

			userCase[userID] = caseType
			userClientType[userID] = clientType
			userDifficulty[userID] = difficulty
			replyMessage(
				req.Events[0].ReplyToken,
				fmt.Sprintf(
					`已建立新案件

案件類型：%s
客戶類型：%s
難度：%s`,
					caseType,
					clientType,
					difficulty,
				),
			)

			return
		}

		aiReply, err := askGroq(
			userID,
			userText,
		)

		if err != nil {
			fmt.Println("groq error:", err)
			aiReply = "系統忙碌中，請稍後再試"
		}

		replyMessage(
			req.Events[0].ReplyToken,
			aiReply,
		)
	}

	w.WriteHeader(http.StatusOK)
}

func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "HOME OK")
}

func main() {
	http.HandleFunc("/", home)
	http.HandleFunc("/webhook", webhook)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("start :", port)

	http.ListenAndServe(":"+port, nil)
}

func replyMessage(replyToken string, text string) {

	channelToken := os.Getenv("LINE_CHANNEL_TOKEN")
	fmt.Println("token length:", len(channelToken))
	payload := map[string]interface{}{
		"replyToken": replyToken,
		"messages": []map[string]string{
			{
				"type": "text",
				"text": text,
			},
		},
	}

	b, _ := json.Marshal(payload)

	req, _ := http.NewRequest(
		"POST",
		"https://api.line.me/v2/bot/message/reply",
		bytes.NewBuffer(b),
	)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+channelToken)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println("reply error:", err)
		return
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	fmt.Println("LINE Reply Status:", resp.StatusCode)
	fmt.Println(string(body))
}
