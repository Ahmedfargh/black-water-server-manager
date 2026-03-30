package NotificationDrivers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type TelegramNotificationDriver struct {
	BotToken string
	ChatID   string
}

func (t *TelegramNotificationDriver) Send(recipient string, message string, metadata map[string]string) (any, error) {
	url := "https://api.telegram.org/bot" + t.BotToken + "/sendMessage"

	payload := map[string]string{
		"chat_id": t.ChatID,
		"text":    message,
	}
	jsonPayload, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("error")
		return "", fmt.Errorf("failed with status: %s", resp.Status)
	}
	return "", nil
}

func (t *TelegramNotificationDriver) GetStatus(notificationID string) (any, error) {

	return "delivered", nil
}

func (t *TelegramNotificationDriver) SetBotToken(BotToken string) {
	t.BotToken = BotToken
}
func (t *TelegramNotificationDriver) SetChatID(ChatID string) {
	t.ChatID = ChatID
}
