// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

// Exports for use in tests only.
var (
	ResourceListenerCertificate = resourceListenerCertificate

	FindListenerByARN                   = findListenerByARN
	FindListenerCertificateByTwoPartKey = findListenerCertificateByTwoPartKey
	HealthCheckProtocolEnumValues       = healthCheckProtocolEnumValues
	HostedZoneIDPerRegionALBMap         = hostedZoneIDPerRegionALBMap
	HostedZoneIDPerRegionNLBMap         = hostedZoneIDPerRegionNLBMap
	ProtocolVersionEnumValues           = protocolVersionEnumValues
)

const (
	AlpnPolicyHTTP1Only      = alpnPolicyHTTP1Only
	AlpnPolicyHTTP2Only      = alpnPolicyHTTP2Only
	AlpnPolicyHTTP2Optional  = alpnPolicyHTTP2Optional
	AlpnPolicyHTTP2Preferred = alpnPolicyHTTP2Preferred
	AlpnPolicyNone           = alpnPolicyNone

	LoadBalancerAttributeClientKeepAliveSeconds = loadBalancerAttributeClientKeepAliveSeconds

	MutualAuthenticationOff         = mutualAuthenticationOff
	MutualAuthenticationVerify      = mutualAuthenticationVerify
	MutualAuthenticationPassthrough = mutualAuthenticationPassthrough
)
