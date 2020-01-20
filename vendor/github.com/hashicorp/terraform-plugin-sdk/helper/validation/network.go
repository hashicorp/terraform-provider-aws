package validation

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// CIDRNetwork returns a SchemaValidateFunc which tests if the provided value
// is of type string, is in valid CIDR network notation, and has significant bits between min and max (inclusive)
func CIDRNetwork(min, max int) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		_, ipnet, err := net.ParseCIDR(v)
		if err != nil {
			es = append(es, fmt.Errorf(
				"expected %s to contain a valid CIDR, got: %s with err: %s", k, v, err))
			return
		}

		if ipnet == nil || v != ipnet.String() {
			es = append(es, fmt.Errorf(
				"expected %s to contain a valid network CIDR, expected %s, got %s",
				k, ipnet, v))
		}

		sigbits, _ := ipnet.Mask.Size()
		if sigbits < min || sigbits > max {
			es = append(es, fmt.Errorf(
				"expected %q to contain a network CIDR with between %d and %d significant bits, got: %d",
				k, min, max, sigbits))
		}

		return
	}
}

// SingleIP returns a SchemaValidateFunc which tests if the provided value
// is of type string, and in valid single IP notation
func SingleIP() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		ip := net.ParseIP(v)
		if ip == nil {
			es = append(es, fmt.Errorf(
				"expected %s to contain a valid IP, got: %s", k, v))
		}
		return
	}
}

// IPRange returns a SchemaValidateFunc which tests if the provided value
// is of type string, and in valid IP range notation
func IPRange() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		ips := strings.Split(v, "-")
		if len(ips) != 2 {
			es = append(es, fmt.Errorf(
				"expected %s to contain a valid IP range, got: %s", k, v))
			return
		}
		ip1 := net.ParseIP(ips[0])
		ip2 := net.ParseIP(ips[1])
		if ip1 == nil || ip2 == nil || bytes.Compare(ip1, ip2) > 0 {
			es = append(es, fmt.Errorf(
				"expected %s to contain a valid IP range, got: %s", k, v))
		}
		return
	}
}
