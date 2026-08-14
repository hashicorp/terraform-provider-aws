// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mailmanager

var (
	ResourceTrafficPolicy = newTrafficPolicyResource
	FindTrafficPolicyByID = findTrafficPolicyByID

	FindRuleSetByID = findRuleSetByID
	ResourceRuleSet = newRuleSetResource

	FindIngressPointByID = findIngressPointByID
	ResourceIngressPoint = newIngressPointResource
)
