// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appmesh

// Exports for use in tests only.
var (
	ResourceGatewayRoute   = resourceGatewayRoute
	ResourceMesh           = resourceMesh
	ResourceRoute          = resourceRoute
	ResourceVirtualGateway = resourceVirtualGateway
	ResourceVirtualNode    = resourceVirtualNode
	ResourceVirtualRouter  = resourceVirtualRouter
	ResourceVirtualService = resourceVirtualService

	FindGatewayRouteByFourPartKey    = findGatewayRouteByFourPartKey
	FindMeshByTwoPartKey             = findMeshByTwoPartKey
	FindRouteByFourPartKey           = findRouteByFourPartKey
	FindVirtualGatewayByThreePartKey = findVirtualGatewayByThreePartKey
	FindVirtualNodeByThreePartKey    = findVirtualNodeByThreePartKey
	FindVirtualRouterByThreePartKey  = findVirtualRouterByThreePartKey
	FindVirtualServiceByThreePartKey = findVirtualServiceByThreePartKey
)
