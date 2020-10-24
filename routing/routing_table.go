package routing

import (
	"container/list"
	"errors"
	"sync/atomic"
	"time"

	"github.com/RoaringBitmap/gocroaring"
	"github.com/rsocket/rsocket-go"
	"github.com/seal90/seal-cloud-rsocket-go/common/metadata"
)

type RoutingTable struct {
	internalRouteID uint32

	internalRouteIDToRouteID map[uint32]string

	tagsToBitmaps map[TagKey]*gocroaring.Bitmap

	routeEntries map[string]*RouteEntry

	// DirectProcessor<RegisteredEvent> registeredEvents
	registeredEvents *list.List

	// FluxSink<RegisteredEvent> registeredEventsSink
}

func NewRoutingTable() RoutingTable {
	return RoutingTable{
		0,
		make(map[uint32]string),
		make(map[TagKey]*gocroaring.Bitmap),
		make(map[string]*RouteEntry),
		list.New(),
	}
}

// func (routingTable *RoutingTable)

func (routingTable *RoutingTable) RegisterByTagsAndRSocket(tagsMetadata metadata.TagsMetadata, rsocket rsocket.RSocket) {
	routingTable.RegisterByRouteEntry(&RouteEntry{rsocket, tagsMetadata, time.Now().Unix()})
}

func (routingTable *RoutingTable) RegisterByRouteEntry(routeEntry *RouteEntry) error {

	routeID := routeEntry.GetRouteID()

	_, ok := routingTable.routeEntries[routeID]

	if ok {
		return errors.New("Route Id already registered: " + routeID)
	}

	internalRouteID := atomic.AddUint32(&routingTable.internalRouteID, 1)
	routingTable.internalRouteIDToRouteID[internalRouteID] = routeID
	routingTable.routeEntries[routeID] = routeEntry

	tags := routeEntry.GetTags()
	for key, val := range tags {
		tagKey := TagKey{Key: key, Value: val}
		bitmap, ok := routingTable.tagsToBitmaps[tagKey]
		if ok {
			bitmap.Add(internalRouteID)
		} else {
			bitmap = gocroaring.New(internalRouteID)
			routingTable.tagsToBitmaps[tagKey] = bitmap
		}
	}
	return nil
}

func (routingTable *RoutingTable) Deregister(tags metadata.TagsMetadata) bool {
	routeID := tags.GetRouteID()

	findByRouteId := metadata.NewTagsMetadata()
	findByRouteId.AddWellKnownKeyTag(metadata.RouteID, routeID)

	found := routingTable.Find(findByRouteId)
	if found.IsEmpty() || found.GetCardinality() > 1 {
		return false
	}
	internalId, err := found.Select(0)
	if err != nil {
		return false
	}
	delete(routingTable.internalRouteIDToRouteID, internalId)
	delete(routingTable.routeEntries, routeID)
	metadataTags := tags.GetTags()
	for key, val := range metadataTags {
		tagKey := TagKey{key, val}
		bitmap, ok := routingTable.tagsToBitmaps[tagKey]
		if ok {
			bitmap.Remove(internalId)
		}
	}
	return true
}

func (routingTable *RoutingTable) Find(tagsMetadata *metadata.TagsMetadata) (found *gocroaring.Bitmap) {
	found = gocroaring.New()
	if nil == tagsMetadata {
		return
	}

	first := true
	for key, val := range tagsMetadata.GetTags() {
		tagKey := TagKey{Key: key, Value: val}
		search, ok := routingTable.tagsToBitmaps[tagKey]
		if ok {
			if first {
				found.Or(search)
				first = false
			} else {
				found.And(search)
			}
		}

	}
	return
}

func (routingTable *RoutingTable) FindRSockets(tagsMetadata *metadata.TagsMetadata) *list.List {
	listRoutes := list.New()
	found := routingTable.Find(tagsMetadata)
	if found.IsEmpty() {
		return listRoutes
	}
	intIterable := found.Iterator()
	for intIterable.HasNext() {
		internalId := intIterable.Next()
		routeId := routingTable.internalRouteIDToRouteID[internalId]
		routeEntry := routingTable.routeEntries[routeId]
		rSocket := routeEntry.GetRSocket()
		listRoutes.PushBack(&RouteRsocketInfo{routeId, rSocket})
	}
	return listRoutes
}

func (routingTable *RoutingTable) FindRouteIds(tagsMetadata *metadata.TagsMetadata) *list.List {
	listRoutes := list.New()
	found := routingTable.Find(tagsMetadata)
	if found.IsEmpty() {
		return listRoutes
	}
	intIterable := found.Iterator()
	for intIterable.HasNext() {
		internalId := intIterable.Next()
		routeId := routingTable.internalRouteIDToRouteID[internalId]
		listRoutes.PushBack(routeId)
	}
	return listRoutes
}

// func (routingTable *RoutingTable) AddListener(consumer RegisteredEvent) Disposable {
// 	return routingTable.registeredEvents.subscribe(consumer)
// }

type RouteRsocketInfo struct {
	routeId string
	rSocket rsocket.RSocket
}

func (routeRsocketInfo *RouteRsocketInfo) GetRouteID() string {
	return routeRsocketInfo.routeId
}

func (routeRsocketInfo *RouteRsocketInfo) GetRSocket() rsocket.RSocket {
	return routeRsocketInfo.rSocket
}

type TagKey struct {
	Key metadata.TagsMetadataKey

	Value string
}

type RouteEntry struct {
	rSocket rsocket.RSocket

	tagsMetadata metadata.TagsMetadata

	timestamp int64
}

func (routeEntry *RouteEntry) GetRSocket() rsocket.RSocket {
	return routeEntry.rSocket
}

func (routeEntry *RouteEntry) GetRouteID() string {
	return routeEntry.tagsMetadata.GetRouteID()
}

func (routeEntry *RouteEntry) GetTags() map[metadata.TagsMetadataKey]string {
	return routeEntry.tagsMetadata.GetTags()
}

type RegisteredEvent struct {
	RouteEntry *RouteEntry
}

func (event *RegisteredEvent) GetRoutingMetadata() metadata.TagsMetadata {
	return event.RouteEntry.tagsMetadata
}

func (event *RegisteredEvent) GetRSocket() rsocket.RSocket {
	return event.RouteEntry.rSocket
}
