package router

import (
	"main.go/router/manage"
	"main.go/router/observe"
)

type RouterGroup struct {
	Manage  manage.ManageRouterGroup
	Observe observe.OberveRouterGroup
}

var RouterGroupApp = new(RouterGroup)
