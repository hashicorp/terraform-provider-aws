// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

// Exports for use in tests only.
var (
	ResourceApplication                        = newApplicationResource
	ResourceApplicationAssignment              = newApplicationAssignmentResource
	ResourceApplicationAssignmentConfiguration = newApplicationAssignmentConfigurationResource
	ResourceApplicationAccessScope             = newApplicationAccessScopeResource
	ResourceTrustedTokenIssuer                 = newTrustedTokenIssuerResource

	FindApplicationByID                        = findApplicationByID
	FindApplicationAssignmentByID              = findApplicationAssignmentByID
	FindApplicationAssignmentConfigurationByID = findApplicationAssignmentConfigurationByID
	FindApplicationAccessScopeByID             = findApplicationAccessScopeByID
	FindTrustedTokenIssuerByARN                = findTrustedTokenIssuerByARN
)
