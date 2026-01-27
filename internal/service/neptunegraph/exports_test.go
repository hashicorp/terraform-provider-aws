// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package neptunegraph

// Exports for use in tests only.
var (
	ResourceGraph                = newGraphResource
	ResourcePrivateGraphEndpoint = newResourcePrivateGraphEndpoint

	FindGraphByID                = findGraphByID
	FindPrivateGraphEndpointByID = findPrivateGraphEndpointByID
)
