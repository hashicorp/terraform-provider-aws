// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

// Exports for use in tests only.
var (
	FindListenerByARN             = findListenerByARN
	HealthCheckProtocolEnumValues = healthCheckProtocolEnumValues
	ProtocolVersionEnumValues     = protocolVersionEnumValues
)

const (
	MutualAuthenticationOff         = mutualAuthenticationOff
	MutualAuthenticationVerify      = mutualAuthenticationVerify
	MutualAuthenticationPassthrough = mutualAuthenticationPassthrough
)

const (
	AlpnPolicyHTTP1Only      = alpnPolicyHTTP1Only
	AlpnPolicyHTTP2Only      = alpnPolicyHTTP2Only
	AlpnPolicyHTTP2Optional  = alpnPolicyHTTP2Optional
	AlpnPolicyHTTP2Preferred = alpnPolicyHTTP2Preferred
	AlpnPolicyNone           = alpnPolicyNone

	LoadBalancerAttributeClientKeepAliveSeconds = loadBalancerAttributeClientKeepAliveSeconds
)
