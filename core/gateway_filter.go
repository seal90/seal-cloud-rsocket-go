package core

type GatewayFilter interface {
	Filter(exchange GatewayExchange, chain GatewayFilterChain)
}
