// Copyright IBM Corp. 2014, 2026
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

	EnvironmentMigrateState = environmentMigrateState
)
