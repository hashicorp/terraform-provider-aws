package aws

import "net"

func reverseLookup(ip string) (name string) {
	if names, err := net.LookupAddr(ip); err == nil && len(names) > 0 {
		name = names[0]
	}
	return
}
