package notification

import (
	"jacobo.tarrio.org/jtweb/comments/engine"
)

type NotificationEngine interface {
	Notify(comment *engine.Comment) error
}

type nullEngine struct{}

func NullNotificationEngine() NotificationEngine {
	return &nullEngine{}
}

func (*nullEngine) Notify(*engine.Comment) error {
	return nil
}
