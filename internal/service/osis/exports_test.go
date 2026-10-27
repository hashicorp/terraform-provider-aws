// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package osis

// Exports for use in tests only.
var (
	FindPipelineByName              = findPipelineByName
	FindPipelineEndpointByID        = findPipelineEndpointByID
	FindResourcePolicyByResourceARN = findResourcePolicyByResourceARN

	ResourcePipeline         = newPipelineResource
	ResourcePipelineEndpoint = newPipelineEndpointResource
	ResourceResourcePolicy   = newResourcePolicyResource
)
