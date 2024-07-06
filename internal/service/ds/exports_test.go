// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

// Exports for use in tests only.
var (
	ResourceDirectory = resourceDirectory

	DirectoryIDValidator           = directoryIDValidator
	DomainWithTrailingDotValidator = domainWithTrailingDotValidator
	FindDirectoryByID              = findDirectoryByID
	TrustPasswordValidator         = trustPasswordValidator
)
