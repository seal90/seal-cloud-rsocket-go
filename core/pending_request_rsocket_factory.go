package core

import (
	"container/list"

	"github.com/seal90/seal-cloud-rsocket-go/common/metadata"
	"github.com/seal90/seal-cloud-rsocket-go/route"
	"github.com/seal90/seal-cloud-rsocket-go/routing"
)

type PendingRequestRSocketFactory struct {
	routingTable      routing.RoutingTable
	routes            route.Routes
	metadataExtractor metadata.MetadataExtractor
}

func NewPendingRequestRSocketFactory(
	routingTable routing.RoutingTable,
	routes route.Routes,
	metadataExtractor metadata.MetadataExtractor) PendingRequestRSocketFactory {
	return PendingRequestRSocketFactory{routingTable, routes, metadataExtractor}
}

func (factory *PendingRequestRSocketFactory) Create(exchange GatewayExchange) *PendingRequestRSocket {
	pending := factory.ConstructPendingRSocket(exchange)
	disposable := factory.routingTable.AddListener(pending.Accept)
	pending.SetSubscriptionDisposable(disposable)
	return pending
}

func (factory *PendingRequestRSocketFactory) ConstructPendingRSocket(exchange GatewayExchange) *PendingRequestRSocket {
	routeFinder := func(registeredEvent routing.RegisteredEvent) *route.Route {
		return factory.GetRoute(registeredEvent, exchange)
	}

	tagsMetadataConsumer := func(tagsMetadata metadata.TagsMetadata) {
		// tags := exchange.GetTags().and("responder.id", tagsMetadata.GetRouteId())
		// exchange.SetTags(tags)
	}
	return NewPendingRequestRSocket(factory.metadataExtractor, routeFinder, tagsMetadataConsumer)
}

func (factory *PendingRequestRSocketFactory) GetRoute(registeredEvent routing.RegisteredEvent, exchange GatewayExchange) *route.Route {
	r := factory.routes.FindRoute(&exchange.GetRoutingMetadata().TagsMetadata)
	return factory.MatchRoute(r, registeredEvent.GetRoutingMetadata())

}

func (factory *PendingRequestRSocketFactory) MatchRoute(r route.Route, tagsMetadata metadata.TagsMetadata) *route.Route {
	routeIds := factory.routingTable.FindRouteIds(&tagsMetadata)
	if Contains(routeIds, r.GetId()) {
		return &r
	}
	return nil
}

func Contains(l *list.List, value string) bool {
	for e := l.Front(); e != nil; e = e.Next() {
		if e.Value == value {
			return true
		}
	}
	return false
}
