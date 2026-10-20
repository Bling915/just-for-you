package companionship

import "context"

type NotificationRecipient struct{ OpenID string }
type NotificationSender interface {
	Send(context.Context, NotificationRecipient, string, map[string]string) error
}
type NoopNotificationSender struct{}

func (NoopNotificationSender) Send(context.Context, NotificationRecipient, string, map[string]string) error {
	return nil
}
