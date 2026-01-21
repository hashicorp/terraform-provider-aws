// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"fmt"
	"net"
)

// ValidateCIDRBlock validates that the specified CIDR block is valid:
// - The CIDR block parses to an IP address and network
// - The CIDR block is the CIDR block for the network
func ValidateCIDRBlock(cidr string) error {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("%q is not a valid CIDR block: %w", cidr, err)
	}

	if !CIDRBlocksEqual(cidr, ipnet.String()) {
		return fmt.Errorf("%q is not a valid CIDR block; did you mean %q?", cidr, ipnet)
	}

	return nil
}

// ValidateIPv4CIDRBlock validates that the specified CIDR block is valid:
// - The CIDR block parses to an IP address and network
// - The IP address is an IPv4 address
// - The CIDR block is the CIDR block for the network
func ValidateIPv4CIDRBlock(cidr string) error {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("%q is not a valid CIDR block: %w", cidr, err)
	}

	ipv4 := ip.To4()
	if ipv4 == nil {
		return fmt.Errorf("%q is not a valid IPv4 CIDR block", cidr)
	}

	if !CIDRBlocksEqual(cidr, ipnet.String()) {
		return fmt.Errorf("%q is not a valid IPv4 CIDR block; did you mean %q?", cidr, ipnet)
	}

	return nil
}

// ValidateIPv6CIDRBlock validates that the specified CIDR block is valid:
// - The CIDR block parses to an IP address and network
// - The IP address is an IPv6 address
// - The CIDR block is the CIDR block for the network
func ValidateIPv6CIDRBlock(cidr string) error {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("%q is not a valid CIDR block: %w", cidr, err)
	}

	ipv4 := ip.To4()
	if ipv4 != nil {
		return fmt.Errorf("%q is not a valid IPv6 CIDR block", cidr)
	}

	if !CIDRBlocksEqual(cidr, ipnet.String()) {
		return fmt.Errorf("%q is not a valid IPv6 CIDR block; did you mean %q?", cidr, ipnet)
	}

	return nil
}

// CIDRBlocksEqual returns whether or not two CIDR blocks are equal:
// - Both CIDR blocks parse to an IP address and network
// - The string representation of the IP addresses are equal
// - The string representation of the networks are equal
// This function is especially useful for IPv6 CIDR blocks which have multiple valid representations.
func CIDRBlocksEqual(cidr1, cidr2 string) bool {
	ip1, ipnet1, err := net.ParseCIDR(cidr1)
	if err != nil {
		return false
	}
	ip2, ipnet2, err := net.ParseCIDR(cidr2)
	if err != nil {
		return false
	}

	return ip2.String() == ip1.String() && ipnet2.String() == ipnet1.String()
}

// CanonicalCIDRBlock returns the canonical representation of a CIDR block.
// This function is especially useful for hash functions for sets which include IPv6 CIDR blocks.
func CanonicalCIDRBlock(cidr string) string {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return cidr
	}

	return ipnet.String()
}

// CIDRBlocksOverlap returns whether the first CIDR block overlaps with (fully or partially contains)
// the second CIDR block. This works for both IPv4 and IPv6 CIDR blocks.
// Returns false if either CIDR block cannot be parsed.
func CIDRBlocksOverlap(cidr1, cidr2 string) bool {
	_, net1, err := net.ParseCIDR(cidr1)
	if err != nil {
		return false
	}

	ip2, net2, err := net.ParseCIDR(cidr2)
	if err != nil {
		return false
	}

	return net1.Contains(ip2) || net1.Contains(getLastIP(net2))
}

// IsIPv4CIDR returns whether a CIDR block is IPv4.
func IsIPv4CIDR(cidr string) bool {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false
	}
	return ipNet.IP.To4() != nil
}

// getLastIP returns the last IP address in a network
func getLastIP(ipnet *net.IPNet) net.IP {
	ip := make(net.IP, len(ipnet.IP))
	copy(ip, ipnet.IP)

	for i := range ip {
		ip[i] |= ^ipnet.Mask[i]
	}

	return ip
}
