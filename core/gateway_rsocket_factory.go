package core

import (
	metrics "github.com/rcrowley/go-metrics"
	"github.com/seal90/seal-cloud-rsocket-go/common/metadata"
	"github.com/seal90/seal-cloud-rsocket-go/configure"
	"github.com/seal90/seal-cloud-rsocket-go/route"
	"github.com/seal90/seal-cloud-rsocket-go/routing"
)

type GatewayRSocketFactory struct {
	routingTable        routing.RoutingTable
	routes              route.Routes
	pendingFactory      PendingRequestRSocketFactory
	loadBalancerFactory routing.LoadBalancerFactory
	meterRegistry       metrics.Registry
	brokerProperties    configure.BrokerProperties
	metadataExtractor   metadata.MetadataExtractor
}

func NewGatewayRSocketFactory(routingTable routing.RoutingTable, routes route.Routes,
	pendingFactory PendingRequestRSocketFactory,
	loadBalancerFactory routing.LoadBalancerFactory, meterRegistry metrics.Registry,
	brokerProperties configure.BrokerProperties, metadataExtractor metadata.MetadataExtractor) *GatewayRSocketFactory {
	return &GatewayRSocketFactory{
		routingTable,
		routes,
		pendingFactory,
		loadBalancerFactory,
		meterRegistry,
		brokerProperties,
		metadataExtractor,
	}
}

func (gatewayRSocketFactory *GatewayRSocketFactory) Create(tagsMetadata *metadata.TagsMetadata) (*GatewayRSocket, error) {
	tagsMetadata.GetRouteID()
	tagsMetadata.GetByWellKnownKey(metadata.ServiceName)

	gatewayRSocket := NewGatewayRSocket(gatewayRSocketFactory.routes,
		gatewayRSocketFactory.pendingFactory, gatewayRSocketFactory.loadBalancerFactory,
		gatewayRSocketFactory.meterRegistry,
		gatewayRSocketFactory.brokerProperties, gatewayRSocketFactory.metadataExtractor, tagsMetadata)
	return gatewayRSocket, nil
}
