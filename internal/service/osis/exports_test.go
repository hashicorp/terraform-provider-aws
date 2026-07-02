// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package osis

// Exports for use in tests only.
var (
	FindPipelineByName       = findPipelineByName
	FindPipelineEndpointByID = findPipelineEndpointByID

	ResourcePipeline               = newPipelineResource
	ResourcePipelineEndpoint       = newPipelineEndpointResource
	ResourcePipelineResourcePolicy = newPipelineResourcePolicyResource
)
