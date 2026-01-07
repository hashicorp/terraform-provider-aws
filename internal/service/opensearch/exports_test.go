// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearch

// Exports for use in tests only.
var (
	ResourceAuthorizeVPCEndpointAccess = newAuthorizeVPCEndpointAccessResource
	ResourceDomainSAMLOptions          = resourceDomainSAMLOptions
	ResourceInboundConnectionAccepter  = resourceInboundConnectionAccepter
	ResourceOutboundConnection         = resourceOutboundConnection
	ResourcePackage                    = resourcePackage
	ResourcePackageAssociation         = resourcePackageAssociation
	ResourceVPCEndpoint                = resourceVPCEndpoint

	FindAuthorizeVPCEndpointAccessByTwoPartKey = findAuthorizeVPCEndpointAccessByTwoPartKey
	FindDomainByName                           = findDomainByName
	FindPackageByID                            = findPackageByID
	FindPackageAssociationByTwoPartKey         = findPackageAssociationByTwoPartKey
	FindVPCEndpointByID                        = findVPCEndpointByID

	EBSVolumeTypePermitsIopsInput       = ebsVolumeTypePermitsIopsInput
	EBSVolumeTypePermitsThroughputInput = ebsVolumeTypePermitsThroughputInput
	ParseEngineVersion                  = parseEngineVersion
	VPCEndpointsError                   = vpcEndpointsError
	WaitForDomainCreation               = waitForDomainCreation
)
