package configure

import (
	"container/list"
	"math/big"
)

type BrokerProperties struct {
	enable         bool
	id             string
	routeID        big.Int
	serviceName    string
	brokers        list.List
	micrometerTags list.List
}
