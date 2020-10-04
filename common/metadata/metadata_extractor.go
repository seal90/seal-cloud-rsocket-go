package metadata

import (
	"github.com/rsocket/rsocket-go/extension"
	"github.com/rsocket/rsocket-go/payload"
)

const RouteSetupMetadata = "message/x.rsocket.routesetup.v0"

const RouteSetupMetadataKey = "routesetup"

const ForwardingMetadata = "message/x.rsocket.forwarding.v0"

const ForwardingMetadataKey = "forwarding"

var extractors map[string]func(string, []byte, map[string]interface{})

func init() {
	extractors = make(map[string]func(string, []byte, map[string]interface{}), 3)
	extractors[extension.MessageRouting.String()] = RoutingExtrace
	extractors[RouteSetupMetadata] = RouteSetupExtrace
	extractors[ForwardingMetadata] = ForwardingExtrace
}

// MetadataExtractor rsockt metadata extractor
type MetadataExtractor interface {
	Extract(p payload.Payload, t string) (map[string]interface{}, error)
}

// DefaultMetadataExtractor default rsockt metadata extractor
type DefaultMetadataExtractor struct {
}

func (d DefaultMetadataExtractor) Extract(data payload.Payload, metadataType string) (map[string]interface{}, error) {
	return Extract(data, metadataType)
}

func Extract(data payload.Payload, metadataType string) (map[string]interface{}, error) {
	result := make(map[string]interface{}, 3)
	metadata, _ := data.Metadata()
	if metadataType == extension.MessageCompositeMetadata.String() {
		compositeMetadata := extension.NewCompositeMetadataBytes(metadata)
		scanner := compositeMetadata.Scanner()
		for scanner.Scan() {
			subMimeType, subMetadata, _ := scanner.Metadata()
			ExtractEntry(subMimeType, subMetadata, result)
		}
	} else {
		ExtractEntry(metadataType, metadata, result)
	}
	return result, nil
}

func ExtractEntry(mimeType string, metadata []byte, result map[string]interface{}) {
	if 0 == len(metadata) {
		return
	}

	extract := extractors[mimeType]
	if nil != extract {
		extract(mimeType, metadata, result)
	}
}

// RoutingExtrace Routing Extrace
func RoutingExtrace(mimeType string, bytes []byte, result map[string]interface{}) {
	tags, _ := extension.ParseRoutingTags(bytes)
	result[mimeType] = tags
}

// RouteSetupExtrace RouteSetup Extrace
func RouteSetupExtrace(mimeType string, bytes []byte, result map[string]interface{}) {
	data := DecodeRouteSetup(bytes)
	result[mimeType] = data
	result[RouteSetupMetadataKey] = data
}

// ForwardingExtrace Forwarding Extrace
func ForwardingExtrace(mimeType string, bytes []byte, result map[string]interface{}) {
	data := DecodeForwarding(bytes)
	result[mimeType] = data
	result[ForwardingMetadataKey] = data
}
