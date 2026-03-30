package Factory

import (
	NotificationDriver "github.com/ahmedfargh/server-manager/Drivers/NotificationDrivers"
)

var drivers = map[string]NotificationDriver.NotificationInterface{
	"Telegram": &NotificationDriver.TelegramNotificationDriver{},
}

type NotificationDriverFactory struct{}

func NewNotificationDriver() *NotificationDriverFactory {
	return &NotificationDriverFactory{}
}

func (f *NotificationDriverFactory) GetDriver(name string, params map[string]any) NotificationDriver.NotificationInterface {
	driver, ok := drivers[name]
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
	}
	return driver
}
