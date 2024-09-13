// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch

// Exports for use in tests only.
var (
	ResourceDomainSAMLOptions         = resourceDomainSAMLOptions
	ResourceInboundConnectionAccepter = resourceInboundConnectionAccepter
	ResourceOutboundConnection        = resourceOutboundConnection
	ResourcePackage                   = resourcePackage
	ResourcePackageAssociation        = resourcePackageAssociation
	ResourceVPCEndpoint               = resourceVPCEndpoint

	FindDomainByName                   = findDomainByName
	FindPackageByID                    = findPackageByID
	FindPackageAssociationByTwoPartKey = findPackageAssociationByTwoPartKey
	FindVPCEndpointByID                = findVPCEndpointByID

	EBSVolumeTypePermitsIopsInput       = ebsVolumeTypePermitsIopsInput
	EBSVolumeTypePermitsThroughputInput = ebsVolumeTypePermitsThroughputInput
	ParseEngineVersion                  = parseEngineVersion
	VPCEndpointsError                   = vpcEndpointsError
	WaitForDomainCreation               = waitForDomainCreation
)
