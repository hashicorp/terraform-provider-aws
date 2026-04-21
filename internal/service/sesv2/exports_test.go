// Copyright IBM Corp. 2014, 2026
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
	ResourceTenant                           = newTenantResource
	ResourceTenantResource                   = newTenantResourceAssociationResource

	FindAccountSuppressionAttributes                 = findAccountSuppressionAttributes
	FindAccountVDMAttributes                         = findAccountVDMAttributes
	FindConfigurationSetByID                         = findConfigurationSetByID
	FindConfigurationSetEventDestinationByTwoPartKey = findConfigurationSetEventDestinationByTwoPartKey
	FindContactListByID                              = findContactListByID
	FindDedicatedIPByTwoPartKey                      = findDedicatedIPByTwoPartKey
	FindDedicatedIPPoolByName                        = findDedicatedIPPoolByName
	FindEmailIdentityByID                            = findEmailIdentityByID
	FindEmailIdentityPolicyByTwoPartKey              = findEmailIdentityPolicyByTwoPartKey
	FindTenantByName                                 = findTenantByName
	FindTenantResourceAssociationByID                = findTenantResourceAssociationByTwoPartKey
)
