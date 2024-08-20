// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

// Exports for use in tests only.
var (
	ResourceClusterEndpoint = resourceClusterEndpoint
	ResourceClusterInstance = resourceClusterInstance

	FindClusterEndpointByTwoPartKey = findClusterEndpointByTwoPartKey
	FindDBInstanceByID              = findDBInstanceByID
)
