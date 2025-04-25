// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

// Exports for use in tests only.
var (
	ResourceAccountAssignment                  = resourceAccountAssignment
	ResourceApplication                        = newApplicationResource
	ResourceApplicationAssignment              = newApplicationAssignmentResource
	ResourceApplicationAssignmentConfiguration = newApplicationAssignmentConfigurationResource
	ResourceApplicationAccessScope             = newApplicationAccessScopeResource
	ResourceCustomerManagedPolicyAttachment    = resourceCustomerManagedPolicyAttachment
	ResourceTrustedTokenIssuer                 = newTrustedTokenIssuerResource

	FindAccountAssignmentBy5PartKey            = findAccountAssignmentBy5PartKey
	FindApplicationByID                        = findApplicationByID
	FindApplicationAssignmentByID              = findApplicationAssignmentByID
	FindApplicationAssignmentConfigurationByID = findApplicationAssignmentConfigurationByID
	FindApplicationAccessScopeByID             = findApplicationAccessScopeByID
	FindCustomerManagedPolicyBy4PartKey        = findCustomerManagedPolicyBy4PartKey
	FindTrustedTokenIssuerByARN                = findTrustedTokenIssuerByARN
)
