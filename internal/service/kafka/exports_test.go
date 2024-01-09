// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

// Exports for use in tests only.
var (
	ResourceConfiguration = resourceConfiguration
	ResourceVPCConnection = resourceVPCConnection

	FindConfigurationByARN = findConfigurationByARN
	FindVPCConnectionByARN = findVPCConnectionByARN
)
