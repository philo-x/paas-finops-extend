package service

import (
	"main.go/service/example"
	"main.go/service/manage"
	"main.go/service/observe"
)

type ServiceGroup struct {
	ExampleServiceGroup example.ServiceGroup
	ManageServiceGroup  manage.ManageServiceGroup
	ObserveServiceGroup observe.ObserveServiceGroup
}

var ServiceGroupApp = new(ServiceGroup)
