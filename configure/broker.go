package configure

type Broker struct {
	host           string
	port           uint16
	connectionType ConnectionType
	wsUri          string
}

type ConnectionType int8

const (
	/** TCP RSocket connection. */
	TCP ConnectionType = iota
	/** WEBSOCKET RSocket connection. */
	WEBSocket
)
