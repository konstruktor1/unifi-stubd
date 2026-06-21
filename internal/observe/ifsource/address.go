package ifsource

import (
	"log"
	"net"
	"strconv"
)

// interfaceAddresses returns the first IPv4 address/netmask and all global IPv6
// CIDR addresses for iface.
func interfaceAddresses(iface *net.Interface) (string, string, []string) {
	addrs, err := iface.Addrs()
	if err != nil {
		log.Printf("read addresses for interface %s: %v", iface.Name, err)
		return "", "", nil
	}
	var ipv4, netmask string
	var ipv6 []string
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		ip := ipNet.IP.To4()
		if ip != nil {
			if ipv4 == "" {
				ipv4 = ip.String()
				netmask = net.IP(ipNet.Mask).String()
			}
			continue
		}
		if cidr := ipv6CIDR(ipNet); cidr != "" {
			ipv6 = append(ipv6, cidr)
		}
	}
	return ipv4, netmask, ipv6
}

func ipv6CIDR(ipNet *net.IPNet) string {
	if ipNet == nil {
		return ""
	}
	ip := ipNet.IP
	if ip == nil || ip.To4() != nil || ip.To16() == nil {
		return ""
	}
	if !ip.IsGlobalUnicast() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return ""
	}
	ones, bits := ipNet.Mask.Size()
	if bits != 128 || ones < 0 {
		return ""
	}
	return ip.String() + "/" + strconv.Itoa(ones)
}
