// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

// Exports for use in tests only.
var (
	ResourceAutoScalingConfigurationVersion = resourceAutoScalingConfigurationVersion
	ResourceConnection                      = resourceConnection
	ResourceCustomDomainAssociation         = resourceCustomDomainAssociation
	ResourceObservabilityConfiguration      = resourceObservabilityConfiguration
	ResourceService                         = resourceService
	ResourceVPCConnector                    = resourceVPCConnector
	ResourceVPCIngressConnection            = resourceVPCIngressConnection

	FindAutoScalingConfigurationByARN          = findAutoScalingConfigurationByARN
	FindConnectionByName                       = findConnectionByName
	FindCustomDomainByTwoPartKey               = findCustomDomainByTwoPartKey
	FindDefaultAutoScalingConfigurationSummary = findDefaultAutoScalingConfigurationSummary
	FindObservabilityConfigurationByARN        = findObservabilityConfigurationByARN
	FindServiceByARN                           = findServiceByARN
	FindVPCConnectorByARN                      = findVPCConnectorByARN
	FindVPCIngressConnectionByARN              = findVPCIngressConnectionByARN
	PutDefaultAutoScalingConfiguration         = putDefaultAutoScalingConfiguration
)
