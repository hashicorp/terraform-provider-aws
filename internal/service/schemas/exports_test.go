// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schemas

// Exports for use in tests only.
var (
	ResourceDiscoverer     = resourceDiscoverer
	ResourceRegistry       = resourceRegistry
	ResourceRegistryPolicy = resourceRegistryPolicy
	ResourceSchema         = resourceSchema

	FindDiscovererByID       = findDiscovererByID
	FindRegistryByName       = findRegistryByName
	FindRegistryPolicyByName = findRegistryPolicyByName
	FindSchemaByTwoPartKey   = findSchemaByTwoPartKey
)
