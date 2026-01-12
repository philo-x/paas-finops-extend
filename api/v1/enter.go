package v1

import (
	"main.go/api/v1/manage"
	"main.go/api/v1/observe"
)

type ApiGroup struct {
	ManageApiGroup  manage.ManageGroup
	ObserveApiGroup observe.ObserveGroup
}

var ApiGroupApp = new(ApiGroup)
