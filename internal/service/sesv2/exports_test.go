// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sesv2

// Exports for use in tests only.
var (
	ResourceAccountVDMAttributes             = resourceAccountVDMAttributes
	ResourceConfigurationSet                 = resourceConfigurationSet
	ResourceConfigurationSetEventDestination = resourceConfigurationSetEventDestination
	ResourceContactList                      = resourceContactList
	ResourceDedicatedIPAssignment            = resourceDedicatedIPAssignment
	ResourceDedicatedIPPool                  = resourceDedicatedIPPool
	ResourceEmailIdentity                    = resourceEmailIdentity
	ResourceEmailIdentityFeedbackAttributes  = resourceEmailIdentityFeedbackAttributes
	ResourceEmailIdentityMailFromAttributes  = resourceEmailIdentityMailFromAttributes
	ResourceEmailIdentityPolicy              = resourceEmailIdentityPolicy

	FindAccountSuppressionAttributes                 = findAccountSuppressionAttributes
	FindAccountVDMAttributes                         = findAccountVDMAttributes
	FindConfigurationSetByID                         = findConfigurationSetByID
	FindConfigurationSetEventDestinationByTwoPartKey = findConfigurationSetEventDestinationByTwoPartKey
	FindContactListByID                              = findContactListByID
	FindDedicatedIPByTwoPartKey                      = findDedicatedIPByTwoPartKey
	FindDedicatedIPPoolByName                        = findDedicatedIPPoolByName
	FindEmailIdentityByID                            = findEmailIdentityByID
	FindEmailIdentityPolicyByTwoPartKey              = findEmailIdentityPolicyByTwoPartKey
)
