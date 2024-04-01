// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

// Exports for use in tests only.
var (
	ResourceBlockPublicAccessConfiguration = resourceBlockPublicAccessConfiguration
	ResourceCluster                        = resourceCluster
	ResourceInstanceFleet                  = resourceInstanceFleet

	FindBlockPublicAccessConfiguration = findBlockPublicAccessConfiguration
	FindClusterByID                    = findClusterByID
	FindInstanceFleetByTwoPartKey      = findInstanceFleetByTwoPartKey
)
