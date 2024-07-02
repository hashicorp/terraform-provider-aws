// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package drs

// Exports for use in tests only.
var (
	ResourceReplicationConfigurationTemplate = newReplicationConfigurationTemplateResource

	FindReplicationConfigurationTemplateByID = findReplicationConfigurationTemplateByID
)
