package Factory

import (
	"fmt"

	NotificationDriver "github.com/ahmedfargh/server-manager/Drivers/NotificationDrivers"
)

var drivers = map[string]NotificationDriver.NotificationInterface{
	"Telegram": &NotificationDriver.TelegramNotificationDriver{},
	"Discord":  &NotificationDriver.DiscordDriver{},
	"Webhook":  &NotificationDriver.WebHookDriver{},
}

type NotificationDriverFactory struct{}

func NewNotificationDriver() *NotificationDriverFactory {
	return &NotificationDriverFactory{}
}

func (f *NotificationDriverFactory) GetDriver(name string, params map[string]any) NotificationDriver.NotificationInterface {
	driver, ok := drivers[name]
	fmt.Println(driver)
	if !ok {
		return nil
	}
	if name == "Telegram" {
		if tg, ok := driver.(*NotificationDriver.TelegramNotificationDriver); ok {
			if token, ok := params["BotToken"].(string); ok {
				tg.SetBotToken(token)
			}
			if chatID, ok := params["ChatID"].(string); ok {
				tg.SetChatID(chatID)
			}
		}
	} else if name == "Discord" {
		if dc, ok := driver.(*NotificationDriver.DiscordDriver); ok {
			fmt.Println("Discord driver Is Iniated")
			if token, ok := params["BotToken"].(string); ok {
				dc.SetBotToken(token)
			}
			if channelID, ok := params["ChannelID"].(string); ok {
				dc.SetChannelID(channelID)
			}
		}
	} else if name == "Webhook" {
		fmt.Println("Webhook driver Is Iniated")
		if wh, ok := driver.(*NotificationDriver.WebHookDriver); ok {
			fmt.Println("Webhook driver Is Iniated")
			if url, ok := params["URL"].(string); ok {
				wh.URL = url
			}
			if secret, ok := params["WebHookSecret"].(string); ok {
				wh.WebHookSecret = secret
			}
		}
	}
	return driver
}
