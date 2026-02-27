package service

import (
	"main.go/service/example"
	"main.go/service/manage"
	"main.go/service/observe"
	"main.go/service/webhook"
)

type ServiceGroup struct {
	ExampleServiceGroup example.ServiceGroup
	ManageServiceGroup  manage.ManageServiceGroup
	ObserveServiceGroup observe.ObserveServiceGroup
	WebhookServiceGroup webhook.ServiceGroup
}

var ServiceGroupApp = new(ServiceGroup)
