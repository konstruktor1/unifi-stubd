package ifsource

import (
	"log"
	"net"
)

// firstInterfaceIPv4 returns the first IPv4 address and netmask for iface.
func firstInterfaceIPv4(iface *net.Interface) (string, string) {
	addrs, err := iface.Addrs()
	if err != nil {
		log.Printf("read addresses for interface %s: %v", iface.Name, err)
		return "", ""
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		ip := ipNet.IP.To4()
		if ip == nil {
			continue
		}
		return ip.String(), net.IP(ipNet.Mask).String()
	}
	return "", ""
}
