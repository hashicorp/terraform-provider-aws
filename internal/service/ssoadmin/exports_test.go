// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

// Exports for use in tests only.
var (
	ResourceApplication                        = newResourceApplication
	ResourceApplicationAssignment              = newResourceApplicationAssignment
	ResourceApplicationAssignmentConfiguration = newResourceApplicationAssignmentConfiguration
	ResourceTrustedTokenIssuer                 = newResourceTrustedTokenIssuer

	FindApplicationByID                        = findApplicationByID
	FindApplicationAssignmentByID              = findApplicationAssignmentByID
	FindApplicationAssignmentConfigurationByID = findApplicationAssignmentConfigurationByID
	FindTrustedTokenIssuerByARN                = findTrustedTokenIssuerByARN
)
