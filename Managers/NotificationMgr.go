package Managers

import (
	"fmt"

	time "time"

	"github.com/ahmedfargh/server-manager/Database/Models"
	"github.com/ahmedfargh/server-manager/Drivers/NotificationDrivers"
	Factory "github.com/ahmedfargh/server-manager/Factory"
)

type NotificationManager struct {
	Driver NotificationDrivers.NotificationInterface
}

func NewNotificationManager(driver NotificationDrivers.NotificationInterface) *NotificationManager {
	return &NotificationManager{
		Driver: driver,
	}
}

func (nm *NotificationManager) Notify(recipient string, message string, metadata map[string]string) (any, error) {
	return nm.Driver.Send(recipient, message, metadata)
}

func (nm *NotificationManager) CheckNotificationStatus(notificationID string) (any, error) {
	return nm.Driver.GetStatus(notificationID)

}
func (nm *NotificationManager) NotifyUsers(users []Models.User, message string, metadata map[string]string) {
	for _, user := range users {

		factory := Factory.NewNotificationDriver()
		params := map[string]any{
			"BotToken": user.TelegramBotToken,
			"ChatID":   user.TelegramChatID,
		}

		drv := factory.GetDriver(user.NotificationDriver, params)
		if drv != nil {
			if user.NotificationDriver == "Telegram" {
				mgr := NewNotificationManager(drv)
				mgr.Notify("", message, metadata)
			} else if user.NotificationDriver == "Discord" {
				fmt.Println("Sending Discord Notification")
				metadata["title"] = "BlackWater Server Manager"
				metadata["description"] = message
				metadata["bot_token"] = user.DiscordBotToken
				metadata["channel_id"] = user.DiscordChannelID
				mgr := NewNotificationManager(drv)
				mgr.Notify("", message, metadata)
			} else if user.NotificationDriver == "Webhook" {
				fmt.Println("Sending WebHook Notification")
				metadata["event"] = "server_resource_monitoring_report"
				metadata["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
				metadata["payload"] = message
				mgr := NewNotificationManager(drv)
				metadata["URL"] = user.WebHookURL
				metadata["WebHookSecret"] = user.WebHookSecret
				result, err := mgr.Notify("", message, metadata)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(result)
			}
		}

	}
}
