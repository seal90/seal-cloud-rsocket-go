package core

import (
	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/rsocket/rsocket-go/rx"
	"github.com/rsocket/rsocket-go/rx/flux"
	"github.com/rsocket/rsocket-go/rx/mono"
	"github.com/seal90/seal-cloud-rsocket-go/common/metadata"
	"github.com/seal90/seal-cloud-rsocket-go/filter"
	"github.com/seal90/seal-cloud-rsocket-go/route"
	"github.com/seal90/seal-cloud-rsocket-go/routing"
)

type PendingRequestRSocket struct {
	routeFinder            func(routing.RegisteredEvent) *route.Route
	metadataExtractor      metadata.MetadataExtractor
	metadataCallback       func(metadata.TagsMetadata)
	rSocketProcessor       rsocket.RSocket
	subscriptionDisposable *routing.RoutingTableRegisteredEventDispose
	route                  *route.Route
	pendingChannel         chan bool
}

func NewPendingRequestRSocket(
	metadataExtractor metadata.MetadataExtractor,
	routeFinder func(routing.RegisteredEvent) *route.Route,
	metadataCallback func(metadata.TagsMetadata)) *PendingRequestRSocket {

	return &PendingRequestRSocket{routeFinder, metadataExtractor, metadataCallback, nil, nil, nil, make(chan bool)}
}

func (pendingRequestRSocket *PendingRequestRSocket) Accept(registeredEvent routing.RegisteredEvent) {
	route := pendingRequestRSocket.routeFinder(registeredEvent)
	pendingRequestRSocket.route = route
	pendingRequestRSocket.metadataCallback(registeredEvent.GetRoutingMetadata())
	pendingRequestRSocket.rSocketProcessor = registeredEvent.GetRSocket()
	if nil != pendingRequestRSocket.subscriptionDisposable {
		pendingRequestRSocket.subscriptionDisposable.Done()
	}
	pendingRequestRSocket.pendingChannel <- false
}

func (pendingRequestRSocket *PendingRequestRSocket) Processor(logCategory string, p payload.Payload) (rsocket.RSocket, filter.Success) {
	<-pendingRequestRSocket.pendingChannel
	exchange := FromPayload(RequestStream, p, pendingRequestRSocket.metadataExtractor)
	exchange.GetAttributes()[RouteAttr] = pendingRequestRSocket.rSocketProcessor
	successFlag := ExecuteFilterChain((*pendingRequestRSocket.route).GetFilters(), *exchange)
	return pendingRequestRSocket.rSocketProcessor, successFlag
}

// FireAndForget is a single one-way message.
func (pendingRequestRSocket *PendingRequestRSocket) FireAndForget(message payload.Payload) {
	rs, _ := pendingRequestRSocket.Processor("pending-request-faf", message)
	rs.FireAndForget(message)
}

// MetadataPush sends asynchronous Metadata frame.
func (pendingRequestRSocket *PendingRequestRSocket) MetadataPush(message payload.Payload) {

}

// RequestResponse request single response.
func (pendingRequestRSocket *PendingRequestRSocket) RequestResponse(message payload.Payload) mono.Mono {
	rs, _ := pendingRequestRSocket.Processor("pending-request-rr", message)
	return rs.RequestResponse(message)
}

// RequestStream request a completable stream.
func (pendingRequestRSocket *PendingRequestRSocket) RequestStream(message payload.Payload) flux.Flux {
	rs, _ := pendingRequestRSocket.Processor("pending-request-rs", message)
	return rs.RequestStream(message)
}

// RequestChannel request a completable stream in both directions.
func (pendingRequestRSocket *PendingRequestRSocket) RequestChannel(messages rx.Publisher) flux.Flux {
	fluxMessage := messages.(flux.Flux)
	return fluxMessage.SwitchOnFirst(func(s flux.Signal, f flux.Flux) flux.Flux {
		p, _ := s.Value()
		rs, _ := pendingRequestRSocket.Processor("pending-request-rc", p)
		return rs.RequestChannel(messages)

	})

}

func (pendingRequestRSocket *PendingRequestRSocket) SetSubscriptionDisposable(subscriptionDisposable *routing.RoutingTableRegisteredEventDispose) {
	pendingRequestRSocket.subscriptionDisposable = subscriptionDisposable
}
