package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type WebhookRequest struct {
	Events []struct {
		ReplyToken string `json:"replyToken"`
		Message    struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"message"`
	} `json:"events"`
}

func webhook(w http.ResponseWriter, r *http.Request) {

	body, _ := io.ReadAll(r.Body)
	fmt.Println("=== WEBHOOK HIT ===")
	fmt.Println(string(body))
	var req WebhookRequest
	json.Unmarshal(body, &req)

	if len(req.Events) > 0 {
		replyMessage(
			req.Events[0].ReplyToken,
			"您好，我是法律諮詢機器人",
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
