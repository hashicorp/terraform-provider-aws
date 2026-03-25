// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3control

// Exports for use in other packages.
var (
	ListTags   = listTagsImpl
	UpdateTags = updateTagsImpl
)
