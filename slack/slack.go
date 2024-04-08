package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

type NotifyParams struct {
	Text string `json:"text"`
}

func Notify(p *NotifyParams) error {
	return NotifyRaw(SlackWebhookURL, p)
}

func NotifyRaw(slackWebhookURL string, p *NotifyParams) error {
	ctx := context.Background()
	reqBody, err := json.Marshal(p)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", slackWebhookURL, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		log.Println(resp.Status, body)
		return errors.New("notify slack: %s: %s")
	}
	return nil
}

var SlackWebhookURL = "https://hooks.slack.com/services/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"
