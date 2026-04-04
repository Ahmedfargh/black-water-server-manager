package NotificationDrivers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type DiscordDriver struct {
	ChannelID string
	BotToken  string
}

func NewDiscordDriver(channelID string, botToken string) *DiscordDriver {
	return &DiscordDriver{
		ChannelID: channelID,
		BotToken:  botToken,
	}
}
func (d *DiscordDriver) SetChannelID(channelID string) {
	d.ChannelID = channelID
}

func (d *DiscordDriver) SetBotToken(botToken string) {
	d.BotToken = botToken
}
func (d *DiscordDriver) Send(recipient string, message string, metadata map[string]string) (any, error) {
	d.ChannelID = metadata["channel_id"]
	d.BotToken = metadata["bot_token"]
	payload := map[string]interface{}{
		"content": message,
		"tts":     false,
		"embeds": []map[string]interface{}{
			{
				"title":       metadata["title"],
				"description": metadata["description"],
				"color":       5814783,
			},
		},
	}
	jsonValue, err := json.Marshal(payload)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("failed to marshal json: %w", err)
	}
	reader := bytes.NewReader(jsonValue)

	url := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", d.ChannelID)
	req, err := http.NewRequest("POST", url, reader)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bot "+d.BotToken)

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		body, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("discord api error: %s", string(body))
	}

	return response, nil
}
func (d *DiscordDriver) GetStatus(notificationID string) (any, error) {
	// Implement the logic to get the status of a sent notification from Discord
	return nil, nil
}
