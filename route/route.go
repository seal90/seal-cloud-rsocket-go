package route

import (
	"container/list"

	"github.com/seal90/seal-cloud-rsocket-go/common/metadata"
	"github.com/seal90/seal-cloud-rsocket-go/support"
)

type Route interface {
	GetId() string

	GetOrder() int

	GetPredicate() support.AsyncPredicate

	GetFilters() *list.List
}

type DefaultRoute struct {
	ID             string
	TargetMetadata metadata.RouteSetup
	Order          int
	Predicate      support.AsyncPredicate
	GatewayFilters list.List
}

func (route *DefaultRoute) GetId() string {
	return route.ID
}

func (route *DefaultRoute) GetOrder() int {
	return route.Order
}

func (route *DefaultRoute) GetPredicate() support.AsyncPredicate {
	return route.Predicate
}

func (route *DefaultRoute) GetFilters() list.List {
	return route.GatewayFilters
}
