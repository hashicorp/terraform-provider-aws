// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice

// Exports for use in tests only.
var (
	ResourceConfigurationRecorder       = resourceConfigurationRecorder
	ResourceConfigurationRecorderStatus = resourceConfigurationRecorderStatus

	FindConfigurationRecorderByName       = findConfigurationRecorderByName
	FindConfigurationRecorderStatusByName = findConfigurationRecorderStatusByName
)
