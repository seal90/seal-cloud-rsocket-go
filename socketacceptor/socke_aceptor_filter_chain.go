package socketacceptor

import (
	"container/list"

	"github.com/rsocket/rsocket-go"
	"github.com/rsocket/rsocket-go/payload"

	"github.com/seal90/seal-cloud-rsocket-go/common/metadata"
	"github.com/seal90/seal-cloud-rsocket-go/filter"
	"github.com/seal90/seal-cloud-rsocket-go/support"
)

type SocketAcceptorFilterChain struct {
	SocketAcceptorFilters *list.List
	CurrentFilter         SocketAcceptorFilter
	Next                  *SocketAcceptorFilterChain
}

func NewSocketAcceptorFilterChain(socketAcceptorFilters *list.List) *SocketAcceptorFilterChain {
	if nil == socketAcceptorFilters {
		return &SocketAcceptorFilterChain{}
	}
	chain := &SocketAcceptorFilterChain{
		SocketAcceptorFilters: socketAcceptorFilters,
		CurrentFilter:         nil,
		Next:                  nil,
	}
	element := socketAcceptorFilters.Back()
	for nil != element {
		chain = &SocketAcceptorFilterChain{
			SocketAcceptorFilters: socketAcceptorFilters,
			CurrentFilter:         element.Value.(SocketAcceptorFilter),
			Next:                  chain,
		}
		element = element.Prev()
	}

	return chain
}

func (chain *SocketAcceptorFilterChain) Filter(exchange *SocketAcceptorExchange) filter.Success {
	if nil != chain.CurrentFilter && nil != chain.Next {
		return chain.CurrentFilter.Filter(exchange, chain.Next)
	}
	return filter.Instance
}

type SocketAcceptorFilter interface {
	Filter(exchange *SocketAcceptorExchange, next *SocketAcceptorFilterChain) filter.Success
	GetOrder() int
}

type SocketAcceptorExchange struct {
	Setup         payload.SetupPayload
	SendingSocket rsocket.CloseableRSocket
	Metadata      metadata.RouteSetup
}

func (socketAcceptorExchange *SocketAcceptorExchange) GetMetadata() metadata.RouteSetup {
	return socketAcceptorExchange.Metadata
}

func (socketAcceptorExchange *SocketAcceptorExchange) GetSendingSocket() rsocket.CloseableRSocket {
	return socketAcceptorExchange.SendingSocket
}

type SocketAcceptorPredicate interface {
	support.AsyncPredicate
	apply(exchange *SocketAcceptorExchange) bool
}

type SocketAcceptorPredicateFilter struct {
	AsyncPredicate *list.List
}

type DefaultSocketAcceptorPredicate struct {
}

func (defaultSocketAcceptorPredicate DefaultSocketAcceptorPredicate) Apply(inter interface{}) bool {
	return true
}

func NewSocketAcceptorPredicateFilter(asyncPredicates *list.List) *SocketAcceptorPredicateFilter {
	if nil == asyncPredicates || 0 == asyncPredicates.Len() {
		asyncPredicates = list.New()
		asyncPredicates.PushBack(DefaultSocketAcceptorPredicate{})
		return &SocketAcceptorPredicateFilter{asyncPredicates}
	}
	// e := asyncPredicates.Front()
	// asyncPredicate := e.Value.(AsyncPredicate)
	// for ; e != nil; e = e.Next() {
	// 	asyncPredicate.And(e.Value.(AsyncPredicate))
	// }
	//TODO 链调用
	return &SocketAcceptorPredicateFilter{asyncPredicates}
}

func (socketAcceptorPredicateFilter *SocketAcceptorPredicateFilter) Filter(exchange *SocketAcceptorExchange, next *SocketAcceptorFilterChain) filter.Success {
	asyncPredicates := socketAcceptorPredicateFilter.AsyncPredicate
	status := true
	for element := asyncPredicates.Front(); nil != element; element = element.Next() {
		result := element.Value.(support.AsyncPredicate).Apply(exchange)
		if !result {
			status = result
			break
		}
	}
	if status {
		return next.Filter(exchange)
	}
	return filter.Mock
}

func (filter *SocketAcceptorPredicateFilter) GetOrder() int {
	return 10000
}
