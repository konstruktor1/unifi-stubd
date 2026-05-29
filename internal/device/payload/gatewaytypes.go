package payload

import "github.com/konstruktor1/unifi-stubd/internal/device"

type gatewayPayload struct {
	basePayload
	gatewayTelemetry
	gatewayTrafficSummary
	IfTable           []gatewayIfRow                 `json:"if_table"`
	NetworkTable      []gatewayNetworkRow            `json:"network_table"`
	ConfigPortTable   []gatewayConfigPortRow         `json:"config_port_table"`
	EthernetTable     []gatewayEthernetTableRow      `json:"ethernet_table"`
	EthernetOverrides []gatewayEthernetOverrideRow   `json:"ethernet_overrides"`
	PortTable         []gatewayPortRow               `json:"port_table"`
	PortStats         []gatewayPortStatsRow          `json:"port_stats"`
	ReportedNetworks  []gatewayReportedNetworkRow    `json:"reported_networks"`
	Wans              []gatewayWANInventoryRow       `json:"wans,omitempty"`
	Uplink            string                         `json:"uplink"`
	UplinkDepth       int                            `json:"uplink_depth"`
	LastUplink        *gatewayLastUplinkRow          `json:"last_uplink"`
	UplinkTable       []gatewayUplinkRow             `json:"uplink_table"`
	UptimeStats       map[string]gatewayWANHealthRow `json:"uptime_stats,omitempty"`
	InternetHealth    *gatewayInternetHealthRow      `json:"internet_health,omitempty"`
	LastWANStatus     map[string]string              `json:"last_wan_status,omitempty"`
	LastWANIP         string                         `json:"last_wan_ip,omitempty"`
	LANIP             string                         `json:"lan_ip,omitempty"`
	HasEth1           bool                           `json:"has_eth1"`
	HasDPI            bool                           `json:"has_dpi"`
	ConfigNetworkWAN  gatewayConfigNetworkRow        `json:"config_network_wan"`
	ConfigNetworkLAN  *gatewayConfigLANRow           `json:"config_network_lan,omitempty"`
	WAN1              *gatewayWANStatusRow           `json:"wan1,omitempty"`
	ConfigNetworkWAN2 *gatewayConfigNetworkRow       `json:"config_network_wan2,omitempty"`
	WAN2              *gatewayWANStatusRow           `json:"wan2,omitempty"`
}

type connectionFields struct {
	Connected      bool                   `json:"connected"`
	LastConnection *gatewayLastConnection `json:"last_connection"`
}

type gatewayLastConnection struct {
	MAC      string `json:"mac"`
	Source   string `json:"source"`
	IP       string `json:"ip,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	Type     string `json:"type,omitempty"`
}

type gatewayLastUplinkRow struct{}

type gatewayNetworkStats struct {
	counterFields
	optionalRateFields
	gatewayRateFields
}

type gatewayTrafficSummary struct {
	Bytes       int64  `json:"bytes,omitempty"`
	RXBytes     int64  `json:"rx_bytes,omitempty"`
	TXBytes     int64  `json:"tx_bytes,omitempty"`
	BytesRate   *int64 `json:"bytes-r,omitempty"`
	RXBytesRate *int64 `json:"rx_bytes-r,omitempty"`
	TXBytesRate *int64 `json:"tx_bytes-r,omitempty"`
	RXRate      *int64 `json:"rx_rate,omitempty"`
	TXRate      *int64 `json:"tx_rate,omitempty"`
}

type gatewayAssignmentFields struct {
	PortConfID          string `json:"portconf_id,omitempty"`
	NetworkConfID       string `json:"networkconf_id,omitempty"`
	NativeNetworkConfID string `json:"native_networkconf_id,omitempty"`
	NetworkName         string `json:"network_name,omitempty"`
	VLAN                int    `json:"vlan,omitempty"`
}

type gatewayPortLinkFields struct {
	Speed     int    `json:"speed"`
	MaxSpeed  int    `json:"max_speed,omitempty"`
	SpeedCaps int    `json:"speed_caps"`
	Media     string `json:"media"`
}

type gatewayEthernetTableRow struct {
	MAC       string `json:"mac"`
	PortIdx   int    `json:"port_idx"`
	NumPort   int    `json:"num_port"`
	Name      string `json:"name"`
	IfName    string `json:"ifname,omitempty"`
	Media     string `json:"media,omitempty"`
	PortPOE   bool   `json:"port_poe"`
	SpeedCaps int    `json:"speed_caps,omitempty"`
}

type gatewayIfRow struct {
	Name           string `json:"name"`
	IfName         string `json:"ifname"`
	Comment        string `json:"comment"`
	PortIdx        int    `json:"port_idx"`
	MAC            string `json:"mac"`
	IP             string `json:"ip"`
	Netmask        string `json:"netmask"`
	NumPort        int    `json:"num_port"`
	Up             bool   `json:"up"`
	Enable         bool   `json:"enable"`
	NetworkGroup   string `json:"networkgroup"`
	FullDuplex     bool   `json:"full_duplex"`
	PhysicalPorts  []int  `json:"physical_ports"`
	VLAN           int    `json:"vlan,omitempty"`
	ManagementVLAN int    `json:"management_vlan,omitempty"`
	linkFields
	counterFields
	optionalRateFields
	gatewayRateFields
	gatewayWANInlineHealth
	gatewayWANUplinkHealthFields
	SourceInterface string `json:"source_interface"`
	connectionFields
}

type gatewayNetworkRow struct {
	Name                 string              `json:"name"`
	IfName               string              `json:"ifname"`
	GatewayInterfaceName string              `json:"gateway_interface_name,omitempty"`
	PortIdx              int                 `json:"port_idx"`
	MAC                  string              `json:"mac"`
	NetworkGroup         string              `json:"networkgroup"`
	IP                   string              `json:"ip"`
	Netmask              string              `json:"netmask"`
	Address              string              `json:"address"`
	Addresses            []string            `json:"addresses"`
	Up                   string              `json:"up"`
	L1Up                 string              `json:"l1up"`
	Autoneg              string              `json:"autoneg"`
	Duplex               string              `json:"duplex"`
	Speed                string              `json:"speed"`
	MaxSpeed             string              `json:"max_speed"`
	MTU                  string              `json:"mtu"`
	Stats                gatewayNetworkStats `json:"stats"`
	gatewayWANInlineHealth
	SourceInterface string `json:"source_interface"`
	connectionFields
	HostTable []gatewayHostRow `json:"host_table,omitempty"`
}

type gatewayConfigNetworkRow struct {
	Type         string        `json:"type"`
	Name         string        `json:"name,omitempty"`
	IfName       string        `json:"ifname,omitempty"`
	PortIdx      int           `json:"port_idx,omitempty"`
	NetworkGroup string        `json:"networkgroup,omitempty"`
	UplinkIfName string        `json:"uplink_ifname,omitempty"`
	IP           *string       `json:"ip,omitempty"`
	Netmask      *string       `json:"netmask,omitempty"`
	Speed        *string       `json:"speed,omitempty"`
	Autoneg      *bool         `json:"autoneg,omitempty"`
	FullDuplex   *bool         `json:"full_duplex,omitempty"`
	DHCPOptions  []emptyObject `json:"dhcp_options"`
}

type gatewayConfigLANRow struct {
	Name                string  `json:"name,omitempty"`
	IfName              string  `json:"ifname,omitempty"`
	PortIdx             int     `json:"port_idx,omitempty"`
	NetworkGroup        string  `json:"networkgroup,omitempty"`
	IP                  *string `json:"ip,omitempty"`
	Netmask             *string `json:"netmask,omitempty"`
	UplinkIfName        string  `json:"uplink_ifname,omitempty"`
	NetworkConfID       string  `json:"networkconf_id,omitempty"`
	NativeNetworkConfID string  `json:"native_networkconf_id,omitempty"`
	NetworkName         string  `json:"network_name,omitempty"`
	DHCPEnabled         bool    `json:"dhcp_enabled"`
	DHCPRangeStart      string  `json:"dhcp_range_start"`
	DHCPRangeStop       string  `json:"dhcp_range_stop"`
	CIDR                string  `json:"cidr"`
	VLAN                int     `json:"vlan"`
}

type gatewayWANStatusRow struct {
	Type         string `json:"type"`
	Name         string `json:"name"`
	IfName       string `json:"ifname"`
	PortIdx      int    `json:"port_idx"`
	NetworkGroup string `json:"networkgroup"`
	Role         string `json:"role"`
	MAC          string `json:"mac"`
	IP           string `json:"ip"`
	Netmask      string `json:"netmask"`
	Address      string `json:"address"`
	Up           bool   `json:"up"`
	Enable       bool   `json:"enable"`
	Uptime       int    `json:"uptime"`
	Latency      int    `json:"latency"`
	UplinkIfName string `json:"uplink_ifname"`
	linkFields
	counterFields
	optionalRateFields
	gatewayRateFields
	SourceInterface string `json:"source_interface"`
}

type gatewayWANInventoryRow struct {
	Enabled   bool   `json:"enabled"`
	Interface string `json:"interface"`
	IPv4      string `json:"ipv4,omitempty"`
	MAC       string `json:"mac"`
	Plugged   bool   `json:"plugged"`
	Port      int    `json:"port"`
	Type      string `json:"type"`
}

type gatewayInternetHealthRow struct {
	Status         string            `json:"status"`
	WANStatus      map[string]string `json:"wan_status"`
	WANIP          string            `json:"wan_ip,omitempty"`
	Netmask        string            `json:"netmask,omitempty"`
	IfName         string            `json:"ifname,omitempty"`
	UplinkIfName   string            `json:"uplink_ifname,omitempty"`
	PortIdx        int               `json:"port_idx,omitempty"`
	Latency        int               `json:"latency"`
	Uptime         float64           `json:"uptime"`
	Availability   float64           `json:"availability"`
	Downtime       int               `json:"downtime"`
	Drops          int               `json:"drops"`
	IsWANUp        bool              `json:"isWanUp"`
	IsWANConnected bool              `json:"isWanConnected"`
}

type gatewayPortRow struct {
	PortIdx      int                    `json:"port_idx"`
	IfName       string                 `json:"ifname"`
	Name         string                 `json:"name"`
	MAC          string                 `json:"mac"`
	IP           string                 `json:"ip,omitempty"`
	NetworkGroup string                 `json:"networkgroup"`
	Role         string                 `json:"role"`
	Type         string                 `json:"type"`
	NumPort      int                    `json:"num_port"`
	Enable       bool                   `json:"enable"`
	Up           bool                   `json:"up"`
	Connected    bool                   `json:"connected"`
	IsUplink     bool                   `json:"is_uplink"`
	OpMode       string                 `json:"op_mode"`
	FullDuplex   bool                   `json:"full_duplex"`
	Autoneg      bool                   `json:"autoneg"`
	FlowctrlRX   bool                   `json:"flowctrl_rx"`
	FlowctrlTX   bool                   `json:"flowctrl_tx"`
	PortPOE      bool                   `json:"port_poe"`
	POEEnable    bool                   `json:"poe_enable"`
	POECaps      int                    `json:"poe_caps"`
	POEClass     string                 `json:"poe_class,omitempty"`
	POEPower     string                 `json:"poe_power,omitempty"`
	MACTable     []device.MacTableEntry `json:"mac_table"`
	RXBroadcast  int                    `json:"rx_broadcast"`
	RXMulticast  int                    `json:"rx_multicast"`
	RXDropped    int                    `json:"rx_dropped"`
	TXBroadcast  int                    `json:"tx_broadcast"`
	TXMulticast  int                    `json:"tx_multicast"`
	TXDropped    int                    `json:"tx_dropped"`
	gatewayAssignmentFields
	gatewayPortLinkFields
	counterFields
	optionalRateFields
	gatewayRateFields
	gatewayWANInlineHealth
	SourceInterface string `json:"source_interface"`
	connectionFields
}

type gatewayPortStatsRow struct {
	PortIdx     int                    `json:"port_idx"`
	Name        string                 `json:"name"`
	IfName      string                 `json:"ifname,omitempty"`
	MAC         string                 `json:"mac"`
	Type        string                 `json:"type"`
	NumPort     int                    `json:"num_port"`
	Media       string                 `json:"media"`
	Speed       int                    `json:"speed"`
	SpeedCaps   int                    `json:"speed_caps"`
	Enable      bool                   `json:"enable"`
	Up          bool                   `json:"up"`
	IsUplink    bool                   `json:"is_uplink"`
	FullDuplex  bool                   `json:"full_duplex"`
	Autoneg     bool                   `json:"autoneg"`
	FlowctrlRX  bool                   `json:"flowctrl_rx"`
	FlowctrlTX  bool                   `json:"flowctrl_tx"`
	PortPOE     bool                   `json:"port_poe"`
	POEEnable   bool                   `json:"poe_enable"`
	POEClass    string                 `json:"poe_class"`
	POEPower    string                 `json:"poe_power"`
	MACTable    []device.MacTableEntry `json:"mac_table,omitempty"`
	RXBroadcast int                    `json:"rx_broadcast"`
	RXMulticast int                    `json:"rx_multicast"`
	RXDropped   int                    `json:"rx_dropped"`
	TXBroadcast int                    `json:"tx_broadcast"`
	TXMulticast int                    `json:"tx_multicast"`
	TXDropped   int                    `json:"tx_dropped"`
	counterFields
	optionalRateFields
	gatewayRateFields
}

type gatewayConfigPortRow struct {
	Name         string `json:"name"`
	IfName       string `json:"ifname"`
	PortIdx      int    `json:"port_idx"`
	NetworkGroup string `json:"networkgroup"`
	Role         string `json:"role"`
	Up           bool   `json:"up"`
	Enable       bool   `json:"enable"`
	IsUplink     bool   `json:"is_uplink"`
	gatewayAssignmentFields
	linkFields
	gatewayWANInlineHealth
	SourceInterface string `json:"source_interface"`
	connectionFields
}

type gatewayEthernetOverrideRow struct {
	Name         string `json:"name"`
	IfName       string `json:"ifname"`
	PortIdx      int    `json:"port_idx"`
	MAC          string `json:"mac"`
	NetworkGroup string `json:"networkgroup"`
	Role         string `json:"role"`
	Up           bool   `json:"up"`
	Enable       bool   `json:"enable"`
	Disabled     bool   `json:"disabled,omitempty"`
	gatewayAssignmentFields
	linkFields
	gatewayWANInlineHealth
	SourceInterface string `json:"source_interface"`
	connectionFields
}

type gatewayHostRow struct {
	MAC        string `json:"mac"`
	Age        int    `json:"age"`
	Authorized bool   `json:"authorized"`
	RXBytes    int64  `json:"rx_bytes"`
	TXBytes    int64  `json:"tx_bytes"`
	RXPackets  int64  `json:"rx_packets"`
	TXPackets  int64  `json:"tx_packets"`
	Uptime     int    `json:"uptime"`
	Hostname   string `json:"hostname,omitempty"`
	IP         string `json:"ip,omitempty"`
	Type       string `json:"type,omitempty"`
	VLAN       int    `json:"vlan,omitempty"`
	Static     bool   `json:"static,omitempty"`
}

type gatewayReportedNetworkRow struct {
	Name            string   `json:"name"`
	IfName          string   `json:"ifname"`
	PortIdx         int      `json:"port_idx"`
	NetworkGroup    string   `json:"networkgroup"`
	Type            string   `json:"type"`
	IP              string   `json:"ip"`
	Netmask         string   `json:"netmask"`
	Address         string   `json:"address"`
	Addresses       []string `json:"addresses"`
	Up              bool     `json:"up"`
	Availability    float64  `json:"availability,omitempty"`
	Latency         int      `json:"latency,omitempty"`
	Downtime        int      `json:"downtime,omitempty"`
	IsWANConnected  bool     `json:"isWanConnected,omitempty"`
	IsWANUp         bool     `json:"isWanUp,omitempty"`
	SourceInterface string   `json:"source_interface"`
	connectionFields
}

type gatewayUplinkRow struct {
	Name         string  `json:"name"`
	IfName       string  `json:"ifname"`
	PortIdx      int     `json:"port_idx"`
	MAC          string  `json:"mac"`
	Type         string  `json:"type"`
	NetworkGroup string  `json:"networkgroup"`
	UplinkIfName string  `json:"uplink_ifname"`
	IP           string  `json:"ip,omitempty"`
	Up           bool    `json:"up"`
	Enable       bool    `json:"enable"`
	FullDuplex   bool    `json:"full_duplex"`
	Availability float64 `json:"availability,omitempty"`
	Latency      int     `json:"latency,omitempty"`
	Downtime     int     `json:"downtime,omitempty"`
	gatewayWANUplinkHealthFields
	IsWANConnected bool `json:"isWanConnected,omitempty"`
	IsWANUp        bool `json:"isWanUp,omitempty"`
	VLAN           int  `json:"vlan,omitempty"`
	ManagementVLAN int  `json:"management_vlan,omitempty"`
	linkFields
	counterFields
	optionalRateFields
	gatewayRateFields
	SourceInterface string `json:"source_interface"`
	connectionFields
}

type gatewayWANUplinkHealthFields struct {
	Uptime           int     `json:"uptime,omitempty"`
	Drops            int     `json:"drops,omitempty"`
	SpeedtestStatus  string  `json:"speedtest_status,omitempty"`
	SpeedtestLastRun int     `json:"speedtest_lastrun,omitempty"`
	SpeedtestPing    int     `json:"speedtest_ping,omitempty"`
	XputUp           float64 `json:"xput_up,omitempty"`
	XputDown         float64 `json:"xput_down,omitempty"`
}

type gatewayWANHealthRow struct {
	NetworkGroup   string  `json:"networkgroup"`
	IfName         string  `json:"ifname"`
	UplinkIfName   string  `json:"uplink_ifname"`
	IP             string  `json:"ip,omitempty"`
	PortIdx        int     `json:"port_idx"`
	Uptime         float64 `json:"uptime"`
	Availability   float64 `json:"availability"`
	Latency        int     `json:"latency"`
	Downtime       int     `json:"downtime"`
	IsWANUp        bool    `json:"isWanUp"`
	IsWANConnected bool    `json:"isWanConnected"`
}

type gatewayWANInlineHealth struct {
	Availability   *float64 `json:"availability,omitempty"`
	Latency        *int     `json:"latency,omitempty"`
	Downtime       *int     `json:"downtime,omitempty"`
	IsWANUp        *bool    `json:"isWanUp,omitempty"`
	IsWANConnected *bool    `json:"isWanConnected,omitempty"`
}

type gatewayWANHealth struct {
	uptimePercent float64
	latencyMS     int
	downtime      int
	up            bool
	connected     bool
}
