package routing

import (
	"container/list"

	"github.com/seal90/seal-cloud-rsocket-go/common/metadata"
	"github.com/seal90/seal-cloud-rsocket-go/route"
	"github.com/seal90/seal-cloud-rsocket-go/support"
)

type RoutingTableRoutes struct {
	routes       map[string]route.Route
	routingTable RoutingTable
}

func NewRoutingTableRoutes(routingTable RoutingTable) RoutingTableRoutes {
	return RoutingTableRoutes{
		make(map[string]route.Route),
		routingTable,
	}
}

func (routingTableRoutes RoutingTableRoutes) GetRoutes() *list.List {
	routeCollection := list.New()
	// for _, route := range routingTableRoutes.routes {
	// routeCollection.PushBack(route)
	// }
	for routeID, _ := range routingTableRoutes.routingTable.routeEntries {
		route := routingTableRoutes.CreateRoute(routeID)
		routeCollection.PushBack(route)
	}
	return routeCollection
}

func (routingTableRoutes RoutingTableRoutes) FindRoute(data *metadata.TagsMetadata) route.Route {
	routes := routingTableRoutes.GetRoutes()
	for e := routes.Front(); nil != e; e = e.Next() {
		r := e.Value.(route.Route)
		if r.GetPredicate().Apply(data) {
			return r
		}
	}
	return nil
}

func (routingTableRoutes *RoutingTableRoutes) Accept(registeredEvent RegisteredEvent) {
	routingMetadata := registeredEvent.GetRoutingMetadata()
	routeId := routingMetadata.GetRouteID()
	registryRoute := routingTableRoutes.CreateRoute(routeId)
	routingTableRoutes.routes[routeId] = registryRoute

}

func (routingTableRoutes *RoutingTableRoutes) CreateRoute(routeID string) route.Route {
	predicate := NewRoutIdPredicate(routingTableRoutes.routingTable, routeID)
	registryRoute := NewRegistryRoute(routeID, predicate)
	return &registryRoute
}

type RoutIdPredicate struct {
	routingTable RoutingTable
	routeId      string
}

func NewRoutIdPredicate(routingTable RoutingTable, routeId string) RoutIdPredicate {
	return RoutIdPredicate{routingTable, routeId}
}

func (routIdPredicate RoutIdPredicate) Apply(mdate interface{}) bool {
	routeIds := routIdPredicate.routingTable.FindRouteIds(mdate.(*metadata.TagsMetadata))
	for e := routeIds.Front(); nil != e; e = e.Next() {
		if e.Value.(string) == routIdPredicate.routeId {
			return true
		}
	}
	return false
}

type RegistryRoute struct {
	id        string
	predicate support.AsyncPredicate
}

func NewRegistryRoute(id string, predicate support.AsyncPredicate) RegistryRoute {
	return RegistryRoute{id, predicate}
}

func (r *RegistryRoute) GetId() string {
	return r.id
}

func (r *RegistryRoute) GetOrder() int {
	return 0
}

func (r *RegistryRoute) GetPredicate() support.AsyncPredicate {
	return r.predicate
}

func (r *RegistryRoute) GetFilters() *list.List {
	return list.New()
}
