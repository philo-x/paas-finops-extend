package observe

import "main.go/service"

type ObserveGroup struct {
	ObserveAlertApi
}

var observeService = service.ServiceGroupApp.ObserveServiceGroup.ObserveAlertService
