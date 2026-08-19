// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mailmanager

var (
	ResourceArchive = newArchiveResource
	FindArchiveByID = findArchiveByID

	ResourceTrafficPolicy = newTrafficPolicyResource
	FindTrafficPolicyByID = findTrafficPolicyByID

	FindRuleSetByID = findRuleSetByID
	ResourceRuleSet = newRuleSetResource

	ResourceRelay        = newRelayResource
	FindRelayByID        = findRelayByID

	FindIngressPointByID = findIngressPointByID
	ResourceIngressPoint = newIngressPointResource
)
