package routing

import (
	"github.com/seal90/seal-cloud-rsocket-go/filter"
	"github.com/seal90/seal-cloud-rsocket-go/socketacceptor"
)

type RoutingTableSocketAcceptorFilter struct {
	routingTable RoutingTable
}

func NewRoutingTableSocketAcceptorFilter(routingTable RoutingTable) RoutingTableSocketAcceptorFilter {
	return RoutingTableSocketAcceptorFilter{routingTable}
}

// Filter(exchange *SocketAcceptorExchange, next *SocketAcceptorFilterChain) filter.Success

func (routingTableSocketAcceptorFilter RoutingTableSocketAcceptorFilter) Filter(exchange *socketacceptor.SocketAcceptorExchange,
	chain *socketacceptor.SocketAcceptorFilterChain) filter.Success {
	routeSetup := exchange.GetMetadata()
	sendingSocket := exchange.GetSendingSocket()
	routingTableSocketAcceptorFilter.routingTable.RegisterByTagsAndRSocket(routeSetup.TagsMetadata, sendingSocket)
	// This is different from the original treatment
	sendingSocket.OnClose(func(e error) {
		routingTableSocketAcceptorFilter.routingTable.Deregister(routeSetup.TagsMetadata)
	})
	return chain.Filter(exchange)
}

func (routingTableSocketAcceptorFilter RoutingTableSocketAcceptorFilter) GetOrder() int {
	return -1000
}
