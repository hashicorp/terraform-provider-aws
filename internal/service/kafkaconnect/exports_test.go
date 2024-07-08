// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafkaconnect

// Exports for use in tests only.
var (
	ResourceConnector    = resourceConnector
	ResourceCustomPlugin = resourceCustomPlugin

	FindConnectorByARN    = findConnectorByARN
	FindCustomPluginByARN = findCustomPluginByARN
)
