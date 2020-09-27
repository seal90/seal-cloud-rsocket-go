package routing

import (
	"container/list"
	"math/rand"
	"sync/atomic"

	"github.com/seal90/seal-cloud-rsocket-go/common/metadata"
)

type LoadBalancerFactory struct {
	routingTable RoutingTable
}

func NewLoadBalancerFactory(routingTable RoutingTable) LoadBalancerFactory {
	return LoadBalancerFactory{routingTable: routingTable}
}
func (loadBalancerFactory *LoadBalancerFactory) Find(tagsMetadata *metadata.TagsMetadata) *list.List {
	return loadBalancerFactory.routingTable.FindRSockets(tagsMetadata)
}

func (loadBalancerFactory *LoadBalancerFactory) Choose(tagsMetadata *metadata.TagsMetadata) *RouteRsocketInfo {
	rSockets := loadBalancerFactory.routingTable.FindRSockets(tagsMetadata)
	loadBalancer := NewRoundRobinLoadBalancer(tagsMetadata)
	return loadBalancer.Apply(rSockets)
}

type LoadBalancer interface {
	Apply(*list.List) *RouteRsocketInfo
}

type RoundRobinLoadBalancer struct {
	tagsMetadata *metadata.TagsMetadata
	// TODO: change loadbalancer impl based on tags
	// TODO: cache loadbalancers based on tags
	position int32
}

func NewRoundRobinLoadBalancer(tagsMetadata *metadata.TagsMetadata) *RoundRobinLoadBalancer {
	return &RoundRobinLoadBalancer{tagsMetadata, rand.Int31n(1024)}
}

func NewRoundRobinLoadBalancerWithPosition(tagsMetadata *metadata.TagsMetadata, position int32) *RoundRobinLoadBalancer {
	return &RoundRobinLoadBalancer{tagsMetadata, position}
}

func (roundRobinLoadBalancer *RoundRobinLoadBalancer) Apply(rSockets *list.List) *RouteRsocketInfo {
	if 0 == rSockets.Len() {
		return nil
	}
	pos := atomic.AddInt32(&roundRobinLoadBalancer.position, 1)

	collectionPost := int(pos) % rSockets.Len()
	elements := rSockets.Front()
	for i := 0; i < collectionPost; i++ {
		elements.Next()
	}
	return elements.Value.(*RouteRsocketInfo)
}
