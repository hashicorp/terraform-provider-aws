// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

// Exports for use in tests only.
var (
	ResourceAutoScalingConfigurationVersion = resourceAutoScalingConfigurationVersion
	ResourceConnection                      = resourceConnection
	ResourceCustomDomainAssociation         = resourceCustomDomainAssociation
	ResourceObservabilityConfiguration      = resourceObservabilityConfiguration
	ResourceVPCConnector                    = resourceVPCConnector
	ResourceVPCIngressConnection            = resourceVPCIngressConnection

	FindAutoScalingConfigurationByARN   = findAutoScalingConfigurationByARN
	FindConnectionByName                = findConnectionByName
	FindCustomDomainByTwoPartKey        = findCustomDomainByTwoPartKey
	FindObservabilityConfigurationByARN = findObservabilityConfigurationByARN
	FindVPCConnectorByARN               = findVPCConnectorByARN
	FindVPCIngressConnectionByARN       = findVPCIngressConnectionByARN
)
