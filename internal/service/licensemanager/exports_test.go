// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package licensemanager

// Exports for use in tests only.
var (
	ResourceAssociation = resourceAssociation
	ResourceGrant       = resourceGrant

	FindAssociationByTwoPartKey = findAssociationByTwoPartKey
	FindGrantByARN              = findGrantByARN
)
