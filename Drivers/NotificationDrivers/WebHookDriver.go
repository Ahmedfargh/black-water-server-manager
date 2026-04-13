package NotificationDrivers

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	w.URL = metadata["URL"]
	w.WebHookSecret = metadata["WebHookSecret"]
	httpClient := &http.Client{}

	payload := w.SetPayload(metadata)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	reader := bytes.NewBuffer(jsonPayload)
	req, err := http.NewRequest("POST", w.URL, reader)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	_, err = httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	fmt.Println("Webhook notification sent successfully")
	return nil, nil
}
func (w *WebHookDriver) GetStatus(notificationID string) (any, error) {
	return "delivered", nil
}
