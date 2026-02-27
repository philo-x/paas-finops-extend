package v1

import (
	"main.go/api/v1/manage"
	"main.go/api/v1/observe"
	"main.go/api/v1/webhook"
)

type ApiGroup struct {
	ManageApiGroup  manage.ManageGroup
	ObserveApiGroup observe.ObserveGroup
	WebhookApiGroup webhook.ApiGroup
}

var ApiGroupApp = new(ApiGroup)
