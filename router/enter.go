package router

import (
	"main.go/router/manage"
	"main.go/router/observe"
	"main.go/router/webhook"
)

type RouterGroup struct {
	Manage  manage.ManageRouterGroup
	Observe observe.ObserveRouterGroup
	Webhook webhook.WebhookRouterGroup
}

var RouterGroupApp = new(RouterGroup)
