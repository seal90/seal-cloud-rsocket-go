package route

import (
	"container/list"

	"github.com/seal90/seal-cloud-rsocket-go/common/metadata"
)

type Routes interface {
	GetRoutes() *list.List

	// FindRoute(exchange core.GatewayExchange)
	FindRoute(data *metadata.TagsMetadata) Route
}
