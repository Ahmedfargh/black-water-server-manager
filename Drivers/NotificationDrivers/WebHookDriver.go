package NotificationDrivers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type WebHookDriver struct {
	URL           string
	WebHookSecret string
}

func NewWebHookDriver() *WebHookDriver {
	return &WebHookDriver{}
}
func (w *WebHookDriver) SetPayload(metadata map[string]string) map[string]any {
	webhook_payload := make(map[string]any)
	webhook_payload["event"] = metadata["event"]
	webhook_payload["timestamp"] = metadata["timestamp"]
	webhook_payload["payload"] = metadata["payload"]
	return webhook_payload
}
func (w *WebHookDriver) Send(recipient string, message string, metadata map[string]string) (any, error) {
	if w.URL == "" {
		return nil, errors.New("webhook URL is not set")
	}
	httpClient := &http.Client{}

	payload := w.SetPayload(metadata)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewBuffer(jsonPayload)
	req, err := http.NewRequest("POST", w.URL, reader)
	if err != nil {
		return nil, err
	}
	_, err = httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
func (w *WebHookDriver) GetStatus(notificationID string) (any, error) {
	return "delivered", nil
}
