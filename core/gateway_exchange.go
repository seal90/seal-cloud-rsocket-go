package core

import (
	"github.com/rsocket/rsocket-go/extension"
	"github.com/rsocket/rsocket-go/payload"
	"github.com/seal90/seal-cloud-rsocket-go/common/metadata"
)

const RouteAttr = "__route_attr_"

const ForwardingKey = "forwarding"

const RouteKey = "route"

var (
	_keyTypes        map[RSocketRequestType]string
	_keyTypesReverse map[string]RSocketRequestType
)

type RSocketRequestType int8

const (
	/**
	* RSocket fire and forget request type.
	 */
	FireAndForget RSocketRequestType = iota

	/**
	* RSocket request channel request type.
	 */
	RequestChannel

	/**
	* RSocket request response request type.
	 */
	RequestResponse

	/**
	* RSocket request stream request type.
	 */
	RequestStream

	MetadataPush
)

func init() {
	_keyTypes = map[RSocketRequestType]string{
		FireAndForget:   "request.fnf",
		RequestChannel:  "request.channel",
		RequestResponse: "request.response",
		RequestStream:   "request.stream",
		MetadataPush:    "metadata.push",
	}
	_keyTypesReverse = make(map[string]RSocketRequestType, len(_keyTypes))
	for k, v := range _keyTypes {
		_keyTypesReverse[v] = k
	}
}

func (t RSocketRequestType) String() string {
	return _keyTypes[t]
}

func ParseRSocketRequestType(str string) (key RSocketRequestType, ok bool) {
	key, ok = _keyTypesReverse[str]
	if !ok {
		key = -1
	}
	return
}

type GatewayExchange struct {
	rSocketRequestType RSocketRequestType
	routingMetadata    *metadata.Forwarding
	attributes         map[string]interface{}
}

func NewGatewayExchange(rSocketRequestType RSocketRequestType, routingMetadata *metadata.Forwarding) GatewayExchange {
	attributes := make(map[string]interface{})
	return GatewayExchange{rSocketRequestType, routingMetadata, attributes}
}

func (gatewayExchange *GatewayExchange) GetAttributes() map[string]interface{} {
	return gatewayExchange.attributes
}

func (gatewayExchange *GatewayExchange) GetRoutingMetadata() *metadata.Forwarding {
	return gatewayExchange.routingMetadata
}

func (gatewayExchange *GatewayExchange) GetRSocketRequestType() RSocketRequestType {
	return gatewayExchange.rSocketRequestType
}

func FromPayload(t RSocketRequestType, p payload.Payload, metadataExtractor metadata.MetadataExtractor) *GatewayExchange {
	if nil == p {
		return nil
	}
	_, ok := p.Metadata()
	if !ok {
		return nil
	}

	metadataMap, _ := metadataExtractor.Extract(p, extension.MessageCompositeMetadata.String())

	exchange := NewGatewayExchange(t, getForwardingMetadata(metadataMap))

	rdata, ok := metadataMap[RouteKey]
	if ok {
		attributes := exchange.GetAttributes()
		attributes["route-metadata"] = rdata
	}

	return &exchange
}

func getForwardingMetadata(metadataMap map[string]interface{}) *metadata.Forwarding {
	data, ok := metadataMap[ForwardingKey]
	if ok {
		val := data.(metadata.Forwarding)
		return &val
	}
	return nil
}
