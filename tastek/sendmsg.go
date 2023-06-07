package tastek

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ah-its-andy/goconf"
	"github.com/rs/zerolog/log"
)

func SendMessage(subject, msg string) error {
	log.Printf("[DEBUG] Sending message subject: %s content: %s", subject, msg)
	if err := sendPushPlus(subject, msg); err != nil {
		return err
	}
	return nil
}

func sendPushPlus(subject, msg string) error {
	//http://www.pushplus.plus/send?token=7dc703f6b5694ecaa36efaa293dc4a6b&title=XXX&content=XXX&template=html
	token := goconf.GetStringOrDefault("pushplus.token", "")
	if len(token) == 0 {
		return fmt.Errorf("pushplus.token cannot be empty")
	}
	reqUrl := "http://www.pushplus.plus/send"
	payload := map[string]any{
		"token":    token,
		"title":    subject,
		"content":  msg,
		"template": "txt",
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", reqUrl, &buf)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("http response: %v", resp.Status)
	}

	return nil
}
