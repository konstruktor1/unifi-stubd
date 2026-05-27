package device

type portCounterKey int

const (
	portCounterRXBytes portCounterKey = iota
	portCounterTXBytes
	portCounterRXPackets
	portCounterTXPackets
	portCounterRXErrors
	portCounterTXErrors
)

// portCounterOverride binds one override counter to its generated port field.
type portCounterOverride struct {
	key portCounterKey
}

func (binding portCounterOverride) get(override PortOverride) int64 {
	switch binding.key {
	case portCounterRXBytes:
		return override.RXBytes
	case portCounterTXBytes:
		return override.TXBytes
	case portCounterRXPackets:
		return override.RXPackets
	case portCounterTXPackets:
		return override.TXPackets
	case portCounterRXErrors:
		return override.RXErrors
	case portCounterTXErrors:
		return override.TXErrors
	default:
		return 0
	}
}

func (binding portCounterOverride) set(port *Port, value int64) {
	switch binding.key {
	case portCounterRXBytes:
		port.RXBytes = value
	case portCounterTXBytes:
		port.TXBytes = value
	case portCounterRXPackets:
		port.RXPackets = value
	case portCounterTXPackets:
		port.TXPackets = value
	case portCounterRXErrors:
		port.RXErrors = value
	case portCounterTXErrors:
		port.TXErrors = value
	}
}
