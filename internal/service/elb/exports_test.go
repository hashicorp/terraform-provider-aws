// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb

// Exports for use in tests only.
var (
	ResourceAppCookieStickinessPolicy = resourceAppCookieStickinessPolicy

	AppCookieStickinessPolicyParseResourceID     = appCookieStickinessPolicyParseResourceID
	FindLoadBalancerByName                       = findLoadBalancerByName
	FindLoadBalancerListenerPolicyByThreePartKey = findLoadBalancerListenerPolicyByThreePartKey
	FindLoadBalancerPolicyByTwoPartKey           = findLoadBalancerPolicyByTwoPartKey
)
