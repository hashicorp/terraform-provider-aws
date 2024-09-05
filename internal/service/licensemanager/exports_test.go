// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package licensemanager

// Exports for use in tests only.
var (
	ResourceAssociation          = resourceAssociation
	ResourceGrant                = resourceGrant
	ResourceGrantAccepter        = resourceGrantAccepter
	ResourceLicenseConfiguration = resourceLicenseConfiguration

	FindAssociationByTwoPartKey   = findAssociationByTwoPartKey
	FindGrantByARN                = findGrantByARN
	FindReceivedGrantByARN        = findReceivedGrantByARN
	FindLicenseConfigurationByARN = findLicenseConfigurationByARN
)
