// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kafkaconnect

// Exports for use in tests only.
var (
	ResourceConnector           = resourceConnector
	ResourceCustomPlugin        = resourceCustomPlugin
	ResourceWorkerConfiguration = resourceWorkerConfiguration

	FindConnectorByARN           = findConnectorByARN
	FindCustomPluginByARN        = findCustomPluginByARN
	FindWorkerConfigurationByARN = findWorkerConfigurationByARN
)
