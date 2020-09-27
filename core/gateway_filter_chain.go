package core

import (
	"container/list"

	"github.com/seal90/seal-cloud-rsocket-go/filter"
)

type GatewayFilterChain struct {
	allFilters    *list.List
	currentFilter *GatewayFilter
	next          *GatewayFilterChain
}

func Create(allFilters *list.List, currentFilter GatewayFilter, next *GatewayFilterChain) GatewayFilterChain {
	return GatewayFilterChain{allFilters, &currentFilter, next}
}

func ExecuteFilterChain(filters *list.List, exchange GatewayExchange) filter.Success {
	chain := initChain(filters)
	gatewayFilterChain := GatewayFilterChain{filters, chain.currentFilter, chain.next}
	return gatewayFilterChain.Filter(exchange)
}

func initChain(filters *list.List) GatewayFilterChain {
	chain := Create(filters, nil, nil)
	for element := filters.Back(); nil != element; element = element.Prev() {
		chain = Create(filters, element.Value.(GatewayFilter), &chain)
	}
	return chain
}

func (chain *GatewayFilterChain) Filter(exchange GatewayExchange) filter.Success {
	if nil != chain.currentFilter && nil != chain.next {
		(*chain.currentFilter).Filter(exchange, *chain)
	}
	return filter.Instance
}
