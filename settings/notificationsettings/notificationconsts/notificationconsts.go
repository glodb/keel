package notificationconsts

type NotificationType int
type MessageSeverity int

const (
	NOTIFICATION_TYPE_FCM      = NotificationType(1)
	NOTIFICATION_TYPE_SMS      = NotificationType(2)
	NOTIFICATION_TYPE_ACTIVITY = NotificationType(3)

	MESSAGE_SEVERITY_INFO    = MessageSeverity(1)
	MESSAGE_SEVERITY_WARNING = MessageSeverity(2)
	MESSAGE_SEVERITY_ALARM   = MessageSeverity(3)
)
