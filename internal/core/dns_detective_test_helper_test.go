package core

import "net"

func parseIPForTest(value string) net.IP {
	return net.ParseIP(value)
}
