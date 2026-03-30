package Managers

import (
	"fmt"

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
		fmt.Println("Driver:-" + user.NotificationDriver)

		factory := Factory.NewNotificationDriver()
		params := map[string]any{
			"BotToken": user.TelegramBotToken,
			"ChatID":   user.TelegramChatID,
		}

		drv := factory.GetDriver(user.NotificationDriver, params)
		if drv != nil {
			mgr := NewNotificationManager(drv)
			mgr.Notify("", message, metadata)
		}

	}
}
