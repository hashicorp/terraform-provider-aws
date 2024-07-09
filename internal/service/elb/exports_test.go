// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb

// Exports for use in tests only.
var (
	ResourceAppCookieStickinessPolicy = resourceAppCookieStickinessPolicy
	ResourceAttachment                = resourceAttachment
	ResourceBackendServerPolicy       = resourceBackendServerPolicy
	ResourceCookieStickinessPolicy    = resourceCookieStickinessPolicy

	AppCookieStickinessPolicyParseResourceID        = appCookieStickinessPolicyParseResourceID
	BackendServerPolicyParseResourceID              = backendServerPolicyParseResourceID
	FindLoadBalancerAttachmentByTwoPartKey          = findLoadBalancerAttachmentByTwoPartKey
	FindLoadBalancerBackendServerPolicyByTwoPartKey = findLoadBalancerBackendServerPolicyByTwoPartKey
	FindLoadBalancerByName                          = findLoadBalancerByName
	FindLoadBalancerListenerPolicyByThreePartKey    = findLoadBalancerListenerPolicyByThreePartKey
	FindLoadBalancerPolicyByTwoPartKey              = findLoadBalancerPolicyByTwoPartKey
	HostedZoneIDPerRegionMap                        = hostedZoneIDPerRegionMap
	LBCookieStickinessPolicyParseResourceID         = lbCookieStickinessPolicyParseResourceID
)
