package main

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"log"

	metrics "github.com/rcrowley/go-metrics"

	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/extension"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/seal90/seal-cloud-rsocket-go/common/metadata"
	"github.com/seal90/seal-cloud-rsocket-go/configure"
	"github.com/seal90/seal-cloud-rsocket-go/core"
	"github.com/seal90/seal-cloud-rsocket-go/routing"
	"github.com/seal90/seal-cloud-rsocket-go/socketacceptor"
)

func main() {

	closeContext := context.Background()
	canelContext, _ := context.WithCancel(closeContext)

	metadataExtractor := metadata.DefaultMetadataExtractor{}

	socketAcceptorPredicateFilter := socketacceptor.NewSocketAcceptorPredicateFilter(list.New())

	routingTable := routing.NewRoutingTable()

	routingTableSocketAcceptorFilter := routing.NewRoutingTableSocketAcceptorFilter(routingTable)

	socketAcceptorFilters := list.New()
	socketAcceptorFilters.PushBack(socketAcceptorPredicateFilter)
	socketAcceptorFilters.PushBack(routingTableSocketAcceptorFilter)

	filterChain := socketacceptor.NewSocketAcceptorFilterChain(socketAcceptorFilters)

	routingTableRoutes := routing.NewRoutingTableRoutes(routingTable)

	pendingFactory := core.NewPendingRequestRSocketFactory(routingTable, routingTableRoutes, metadataExtractor)

	loadBalancerFactory := routing.NewLoadBalancerFactory(routingTable)

	meterRegistry := metrics.NewRegistry()

	brokerProperties := new(configure.BrokerProperties)

	gatewayRSocketFactory := core.NewGatewayRSocketFactory(routingTable, routingTableRoutes, pendingFactory,
		loadBalancerFactory, meterRegistry, *brokerProperties, metadataExtractor)
	// go func() {
	err := rsocket.Receive().
		OnStart(func() {
			log.Println("============start==============")
		}).
		Resume().
		Fragment(1024).
		Acceptor(func(setup payload.SetupPayload, sendingSocket rsocket.CloseableRSocket) (rsocket.RSocket, error) {

			mimeType := setup.MetadataMimeType()
			mime, ok := extension.ParseMIME(mimeType)
			if !ok {
				return nil, errors.New("not support mime type: " + mimeType)
			}
			metadataMap, _ := metadataExtractor.Extract(setup, mime.String())

			routeSetupMetadata := metadataMap[metadata.RouteSetupMetadata]
			var exchange socketacceptor.SocketAcceptorExchange
			if nil != routeSetupMetadata {
				exchange = socketacceptor.SocketAcceptorExchange{
					Setup:         setup,
					SendingSocket: sendingSocket,
					Metadata:      routeSetupMetadata.(metadata.RouteSetup),
				}
			} else {
				exchange = socketacceptor.SocketAcceptorExchange{
					Setup:         setup,
					SendingSocket: sendingSocket,
					Metadata:      metadata.RouteSetup{},
				}
			}
			doFilterResult := filterChain.Filter(&exchange)
			fmt.Println(doFilterResult)

			tags := exchange.GetMetadata().GetEnrichedTagsMetadata()
			// bind responder
			return gatewayRSocketFactory.Create(&tags)
			// return rsocket.NewAbstractSocket(
			// 	rsocket.RequestResponse(func(msg payload.Payload) mono.Mono {
			// 		log.Println("response:", msg)
			// 		return mono.Just(msg)
			// 	}),
			// ), nil
		}).
		Transport(rsocket.TCPServer().SetAddr(":7002").Build()).
		// Transport("tcp://127.0.0.1:7002").
		Serve(canelContext)
	log.Println(err)
	panic(err)
	// }()
	// time.Sleep(time.Duration(10) * time.Second)

	// closeChan := canelContext.Done()

	// if nil == closeChan {
	// 	log.Println("============start  1==============")
	// 	errInfo := canelContext.Err()
	// 	log.Println(errInfo)
	// } else {
	// 	log.Println("============start  2==============")
	// 	log.Println("============start  3==============")
	// }
}
