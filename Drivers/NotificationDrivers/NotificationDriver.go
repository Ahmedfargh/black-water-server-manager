package NotificationDrivers

type NotificationInterface interface {
	Send(recipient string, message string, metadata map[string]string) (any, error)
	GetStatus(notificationID string) (any, error)
}
