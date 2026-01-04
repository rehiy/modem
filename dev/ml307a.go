package dev

import (
	"github.com/rehiy/modem/at"
)

type ML307A struct {
	CommandSet      *at.CommandSet
	ResponseSet     *at.ResponseSet
	NotificationSet *at.NotificationSet
}

func NewML307A() *ML307A {
	commandSet := at.DefaultCommandSet()
	responseSet := at.DefaultResponseSet()
	notificationSet := at.DefaultNotificationSet()

	return &ML307A{
		CommandSet:      commandSet,
		ResponseSet:     responseSet,
		NotificationSet: notificationSet,
	}
}
