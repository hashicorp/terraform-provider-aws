// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2

// Exports for use in tests only.
var (
	ResourceAccount                           = resourceAccount
	ResourceClassificationExportConfiguration = resourceClassificationExportConfiguration
	ResourceClassificationJob                 = resourceClassificationJob
	ResourceCustomDataIdentifier              = resourceCustomDataIdentifier
	ResourceFindingsFilter                    = resourceFindingsFilter
	ResourceInvitationAccepter                = resourceInvitationAccepter
	ResourceMember                            = resourceMember
	ResourceOrganizationAdminAccount          = resourceOrganizationAdminAccount

	FindClassificationJobByID    = findClassificationJobByID
	FindCustomDataIdentifierByID = findCustomDataIdentifierByID
	FindFindingsFilterByID       = findFindingsFilterByID
	FindMemberByID               = findMemberByID
)
