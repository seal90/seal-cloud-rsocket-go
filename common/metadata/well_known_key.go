package metadata

var (
	_keyTypes        map[WellKnownKey]string
	_keyTypesReverse map[string]WellKnownKey
)

type WellKnownKey int8

const (
	UnparseableKey WellKnownKey = iota - 2
	UnknownReservedKey
	NoTag
	ServiceName
	RouteID
	InstanceName
	ClusterName
	Provider
	Region
	Zone
	Device
	OS
	UserName
	UserID
	MajorVersion
	MinorVersion
	PatchVersion
	Version
	Environment
	TestcEll
	DNS
	IPV4
	IPV6
	Country
	TimeZone       WellKnownKey = 0x1A
	ShardKey       WellKnownKey = 0x1B
	ShardMethod    WellKnownKey = 0x1C
	StickyRouteKey WellKnownKey = 0x1D
	LbMethod       WellKnownKey = 0x1E
	// TODO mabe here bug
	BrokerExtension    WellKnownKey = 0x1F
	WellKnownExtension WellKnownKey = 0x20
)

func init() {
	_keyTypes = map[WellKnownKey]string{
		UnparseableKey:     "UNPARSEABLE_KEY_DO_NOT_USE",
		UnknownReservedKey: "UNKNOWN_YET_RESERVED_DO_NOT_USE",
		NoTag:              "NO_TAG_DO_NOT_USE",
		ServiceName:        "io.rsocket.routing.ServiceName",
		RouteID:            "io.rsocket.routing.RouteId",
		InstanceName:       "io.rsocket.routing.InstanceName",
		ClusterName:        "io.rsocket.routing.ClusterName",
		Provider:           "io.rsocket.routing.Provider",
		Region:             "io.rsocket.routing.Region",
		Zone:               "io.rsocket.routing.Zone",
		Device:             "io.rsocket.routing.Device",
		OS:                 "io.rsocket.routing.OS",
		UserName:           "io.rsocket.routing.UserName",
		UserID:             "io.rsocket.routing.UserId",
		MajorVersion:       "io.rsocket.routing.MajorVersion",
		MinorVersion:       "io.rsocket.routing.MinorVersion",
		PatchVersion:       "io.rsocket.routing.PatchVersion",
		Version:            "io.rsocket.routing.Version",
		Environment:        "io.rsocket.routing.Environment",
		TestcEll:           "io.rsocket.routing.TestCell",
		DNS:                "io.rsocket.routing.DNS",
		IPV4:               "io.rsocket.routing.IPv4",
		IPV6:               "io.rsocket.routing.IPv6",
		Country:            "io.rsocket.routing.Country",
		TimeZone:           "io.rsocket.routing.TimeZone",
		ShardKey:           "io.rsocket.routing.ShardKey",
		ShardMethod:        "io.rsocket.routing.ShardMethod",
		StickyRouteKey:     "io.rsocket.routing.StickyRouteKey",
		LbMethod:           "io.rsocket.routing.LBMethod",
		BrokerExtension:    "Broker Implementation Extension Key",
		WellKnownExtension: "Well Known Extension Key",
	}

	_keyTypesReverse = make(map[string]WellKnownKey, len(_keyTypes))
	for k, v := range _keyTypes {
		_keyTypesReverse[v] = k
	}
}

func (p WellKnownKey) String() string {
	return _keyTypes[p]
}

// ParseWellKnownKey parse a string to WellKnownKey.
func ParseWellKnownKey(str string) (key WellKnownKey, ok bool) {
	key, ok = _keyTypesReverse[str]
	if !ok {
		key = -1
	}
	return
}
