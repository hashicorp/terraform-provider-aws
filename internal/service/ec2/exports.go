// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

// Exports for use in other modules.
var (
	CustomFiltersBlock                                             = customFiltersBlock
	DeleteNetworkInterface                                         = deleteNetworkInterface
	DetachNetworkInterface                                         = detachNetworkInterface
	FindImageByID                                                  = findImageByID
	FindInstanceByID                                               = findInstanceByID
	FindNetworkInterfaces                                          = findNetworkInterfaces
	FindNetworkInterfacesByAttachmentInstanceOwnerIDAndDescription = findNetworkInterfacesByAttachmentInstanceOwnerIDAndDescription
	FindSecurityGroupByDescriptionAndVPCID                         = findSecurityGroupByDescriptionAndVPCID
	FindSecurityGroupByNameAndVPCID                                = findSecurityGroupByNameAndVPCID
	FindSecurityGroupByNameAndVPCIDAndOwnerID                      = findSecurityGroupByNameAndVPCIDAndOwnerID
	FindSecurityGroups                                             = findSecurityGroups
	FindSubnetByID                                                 = findSubnetByID
	FindVPCByID                                                    = findVPCByID
	FindVPCEndpointByID                                            = findVPCEndpointByID
	NetworkInterfaceDetachedTimeout                                = networkInterfaceDetachedTimeout
	NewCustomFilterListFramework                                   = newCustomFilterListFramework
	NewFilter                                                      = newFilter
	ResourceAMI                                                    = resourceAMI
	ResourceSecurityGroup                                          = resourceSecurityGroup
	ResourceTransitGateway                                         = resourceTransitGateway
	ResourceTransitGatewayConnectPeer                              = resourceTransitGatewayConnectPeer
	ResourceVPC                                                    = resourceVPC
	VPCEndpointCreationTimeout                                     = vpcEndpointCreationTimeout
	WaitVPCEndpointAvailable                                       = waitVPCEndpointAvailable
)
