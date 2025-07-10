// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk

// Exports for use in tests only.
var (
	ResourceApplication           = resourceApplication
	ResourceApplicationVersion    = resourceApplicationVersion
	ResourceConfigurationTemplate = resourceConfigurationTemplate
	ResourceEnvironment           = resourceEnvironment

	FindApplicationByName                 = findApplicationByName
	FindApplicationVersionByTwoPartKey    = findApplicationVersionByTwoPartKey
	FindConfigurationSettingsByTwoPartKey = findConfigurationSettingsByTwoPartKey
	FindEnvironmentByID                   = findEnvironmentByID
	HostedZoneIDs                         = hostedZoneIDs
)
