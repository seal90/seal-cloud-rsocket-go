package metadata

import (
	"github.com/rsocket/rsocket-go/extension"
	"github.com/rsocket/rsocket-go/payload"
)

const RouteSetupMetadata = "message/x.rsocket.routesetup.v0"

const RouteSetupMetadataKey = "routesetup"

const ForwardingMetadata = "message/x.rsocket.forwarding.v0"

const ForwardingMetadataKey = "forwarding"

type MetadataExtractor interface {
	Extract(p payload.Payload, t string) (map[string]interface{}, error)
}

type DefaultMetadataExtractor struct {
}

func (extractor DefaultMetadataExtractor) Extract(p payload.Payload, t string) (map[string]interface{}, error) {
	return Extract(p, t)
}

func Extract(payload payload.Payload, metadataMimeType string) (map[string]interface{}, error) {
	result := make(map[string]interface{}, 3)
	metadata, _ := payload.Metadata()
	if metadataMimeType == extension.MessageCompositeMetadata.String() {
		compositeMetadata := extension.NewCompositeMetadataBytes(metadata)
		scanner := compositeMetadata.Scanner()
		for scanner.Scan() {
			subMimeType, subMetadata, _ := scanner.Metadata()
			ExtractEntry(subMimeType, subMetadata, result)
		}
	} else {
		ExtractEntry(metadataMimeType, metadata, result)
	}
	return result, nil
}

func ExtractEntry(mimeType string, metadata []byte, result map[string]interface{}) {
	if 0 == len(metadata) {
		return
	}

	// TODO here
	extractors := make(map[string]func(string, []byte, map[string]interface{}), 3)
	extractors[extension.MessageRouting.String()] = RoutingExtrace
	extractors[RouteSetupMetadata] = RouteSetupExtrace
	extractors[ForwardingMetadata] = ForwardingExtrace

	extract := extractors[mimeType]
	if nil != extract {
		extract(mimeType, metadata, result)
	}
}

func RoutingExtrace(mimeType string, bytes []byte, result map[string]interface{}) {
	tags, _ := extension.ParseRoutingTags(bytes)
	result[mimeType] = tags
}

func RouteSetupExtrace(mimeType string, bytes []byte, result map[string]interface{}) {
	data := DecodeRouteSetup(bytes)
	result[mimeType] = data
	result[RouteSetupMetadataKey] = data
}

func ForwardingExtrace(mimeType string, bytes []byte, result map[string]interface{}) {
	data := DecodeForwarding(bytes)
	result[mimeType] = data
	result[ForwardingMetadataKey] = data
}
