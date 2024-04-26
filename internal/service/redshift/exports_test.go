// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

// Exports for use in tests only.
var (
	ResourceAuthenticationProfile        = resourceAuthenticationProfile
	ResourceClusterIAMRoles              = resourceClusterIAMRoles
	ResourceDataShareAuthorization       = newResourceDataShareAuthorization
	ResourceDataShareConsumerAssociation = newResourceDataShareConsumerAssociation
	ResourceLogging                      = newResourceLogging
	ResourceSnapshotCopy                 = newResourceSnapshotCopy

	FindAuthenticationProfileByID        = findAuthenticationProfileByID
	FindDataShareAuthorizationByID       = findDataShareAuthorizationByID
	FindDataShareConsumerAssociationByID = findDataShareConsumerAssociationByID
	FindLoggingByID                      = findLoggingByID
	FindSnapshotCopyByID                 = findSnapshotCopyByID
)
