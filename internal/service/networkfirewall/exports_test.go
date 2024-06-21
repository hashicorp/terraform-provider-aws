// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

// Exports for use in tests only.
var (
	ResourceLoggingConfiguration       = resourceLoggingConfiguration
	ResourceTLSInspectionConfiguration = newTLSInspectionConfigurationResource

	FindLoggingConfigurationByARN       = findLoggingConfigurationByARN
	FindTLSInspectionConfigurationByARN = findTLSInspectionConfigurationByARN
)
