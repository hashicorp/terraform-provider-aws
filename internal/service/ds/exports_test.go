// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

// Exports for use in tests only.
var (
	ResourceConditionalForwarder = resourceConditionalForwarder
	ResourceDirectory            = resourceDirectory

	DirectoryIDValidator                 = directoryIDValidator
	DomainWithTrailingDotValidator       = domainWithTrailingDotValidator
	FindConditionalForwarderByTwoPartKey = findConditionalForwarderByTwoPartKey
	FindDirectoryByID                    = findDirectoryByID
	TrustPasswordValidator               = trustPasswordValidator
)
