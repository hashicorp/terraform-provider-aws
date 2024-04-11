// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

// Exports for use in tests only.
var (
	ResourceDataShareAuthorization       = newResourceDataShareAuthorization
	ResourceDataShareConsumerAssociation = newResourceDataShareConsumerAssociation
	ResourceLogging                      = newResourceLogging
	ResourceSnapshotCopy                 = newResourceSnapshotCopy

	FindDataShareAuthorizationByID       = findDataShareAuthorizationByID
	FindDataShareConsumerAssociationByID = findDataShareConsumerAssociationByID
	FindLoggingByID                      = findLoggingByID
	FindSnapshotCopyByID                 = findSnapshotCopyByID
)
