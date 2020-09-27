package core

import (
	"container/list"

	metrics "github.com/rcrowley/go-metrics"
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx"
	"github.com/rsocket/rsocket-go/rx/flux"
	"github.com/rsocket/rsocket-go/rx/mono"
	"github.com/seal90/seal-cloud-rsocket-go/common/metadata"
	"github.com/seal90/seal-cloud-rsocket-go/configure"
	"github.com/seal90/seal-cloud-rsocket-go/route"
	"github.com/seal90/seal-cloud-rsocket-go/routing"
)

type GatewayRSocket struct {
	routes              route.Routes
	pendingFactory      PendingRequestRSocketFactory
	loadBalancerFactory routing.LoadBalancerFactory
	meterRegistry       metrics.Registry
	properties          configure.BrokerProperties
	metadataExtractor   metadata.MetadataExtractor
	tagsMetadata        *metadata.TagsMetadata
}

func NewGatewayRSocket(
	routes route.Routes,
	pendingFactory PendingRequestRSocketFactory,
	loadBalancerFactory routing.LoadBalancerFactory,
	meterRegistry metrics.Registry,
	properties configure.BrokerProperties,
	metadataExtractor metadata.MetadataExtractor,
	tagsMetadata *metadata.TagsMetadata) *GatewayRSocket {

	return &GatewayRSocket{routes, pendingFactory, loadBalancerFactory, meterRegistry, properties, metadataExtractor, tagsMetadata}
}

// FireAndForget is a single one-way message.
func (gatewayRSocket *GatewayRSocket) FireAndForget(msg payload.Payload) {
	exchange := gatewayRSocket.CreateExchange(FireAndForget, msg)
	rSockets := gatewayRSocket.FindRSocketOrCreatePending(*exchange)
	for rs := rSockets.Front(); nil != rs; rs = rs.Next() {
		rs.Value.(rsocket.RSocket).FireAndForget(msg)
	}
}

// MetadataPush sends asynchronous Metadata frame.
func (gatewayRSocket *GatewayRSocket) MetadataPush(msg payload.Payload) {
	exchange := gatewayRSocket.CreateExchange(MetadataPush, msg)
	rSockets := gatewayRSocket.FindRSocketOrCreatePending(*exchange)
	for rs := rSockets.Front(); nil != rs; rs = rs.Next() {
		rs.Value.(rsocket.RSocket).MetadataPush(msg)
	}
}

// RequestResponse request single response.
func (gatewayRSocket *GatewayRSocket) RequestResponse(msg payload.Payload) mono.Mono {
	exchange := gatewayRSocket.CreateExchange(RequestResponse, msg)
	rSockets := gatewayRSocket.FindRSocketOrCreatePending(*exchange)
	for rs := rSockets.Front(); nil != rs; rs = rs.Next() {
		s := rs.Value.(*rsocket.RSocket)
		ss := *s
		return ss.RequestResponse(msg)
	}
	return mono.Empty()
}

// RequestStream request a completable stream.
func (gatewayRSocket *GatewayRSocket) RequestStream(msg payload.Payload) flux.Flux {
	exchange := gatewayRSocket.CreateExchange(RequestStream, msg)
	rSockets := gatewayRSocket.FindRSocketOrCreatePending(*exchange)
	for rs := rSockets.Front(); nil != rs; rs = rs.Next() {
		return rs.Value.(rsocket.RSocket).RequestStream(msg)
	}
	return flux.Empty()
}

// RequestChannel request a completable stream in both directions.
func (gatewayRSocket *GatewayRSocket) RequestChannel(msgs rx.Publisher) flux.Flux {
	fluxMsgs := msgs.(flux.Flux)
	return fluxMsgs.SwitchOnFirst(func(s flux.Signal, f flux.Flux) flux.Flux {
		p, _ := s.Value()
		exchange := gatewayRSocket.CreateExchange(RequestChannel, p)
		rSockets := gatewayRSocket.FindRSocketOrCreatePending(*exchange)
		for rs := rSockets.Front(); nil != rs; rs = rs.Next() {
			return rs.Value.(rsocket.RSocket).RequestChannel(msgs)
		}
		return flux.Empty()
	})

}

func (gatewayRSocket *GatewayRSocket) CreateExchange(t RSocketRequestType, p payload.Payload) *GatewayExchange {
	exchange := FromPayload(t, p, gatewayRSocket.metadataExtractor)
	// tags := getTags(exchange)
	// exchange.setTags(tags)
	return exchange
}

func getTags(exchange GatewayExchange) {
	// forwarding := exchange.GetRoutingMetadata()
	// requesterName:=forwarding.GetTags()[metadata.TagsMetadataKey{metadata.ServiceName, ""}]
	// requesterId := forwarding.GetRouteID()
	// responderName := exchange.GetRoutingMetadata().get
	// String responderName = "FIXME"; // FIXME: exchange.getRoutingMetadata().getName();
	// metrics.T
}

func (gatewayRSocket *GatewayRSocket) FindRSocketOrCreatePending(exchange GatewayExchange) *list.List {
	forwarding := exchange.GetRoutingMetadata()
	route := gatewayRSocket.routes.FindRoute(&forwarding.TagsMetadata)
	if nil != route {
		exchange.GetAttributes()[RouteAttr] = route
		return gatewayRSocket.findRSocketOrCreatePending(exchange, route)
	}
	return gatewayRSocket.CreatePending(exchange)
}

func (gatewayRSocket *GatewayRSocket) CreatePending(exchange GatewayExchange) *list.List {
	rSockets := list.New()
	pending := gatewayRSocket.pendingFactory.Create(exchange)
	rSockets.PushBack(pending)
	return rSockets
}

func (gatewayRSocket *GatewayRSocket) findRSocketOrCreatePending(exchange GatewayExchange, route route.Route) *list.List {
	ExecuteFilterChain(route.GetFilters(), exchange)
	tags := exchange.GetRoutingMetadata().GetTags()
	_, ok := tags[metadata.TagsMetadataKey{metadata.WellKnownKey(0), "multicast"}]
	var routeRsocketInfos *list.List
	if ok {
		routeRsocketInfos = gatewayRSocket.loadBalancerFactory.Find(&exchange.GetRoutingMetadata().TagsMetadata)
	} else {
		routeRsocketInfo := gatewayRSocket.loadBalancerFactory.Choose(&exchange.GetRoutingMetadata().TagsMetadata)
		routeRsocketInfos = list.New()
		routeRsocketInfos.PushBack(routeRsocketInfo)
	}

	rsockets := list.New()
	for routeRsocketInfo := routeRsocketInfos.Front(); nil != routeRsocketInfo; routeRsocketInfo = routeRsocketInfo.Next() {
		info := routeRsocketInfo.Value.(*routing.RouteRsocketInfo)
		rsockets.PushBack(info.GetRSocket())
	}
	return rsockets
}
