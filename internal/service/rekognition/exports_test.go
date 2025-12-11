// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package rekognition

// Exports for use in tests only.

var (
	ResourceProject         = newProjectResource
	ResourceCollection      = newCollectionResource
	ResourceStreamProcessor = newStreamProcessorResource
)

var (
	FindCollectionByID        = findCollectionByID
	FindProjectByName         = findProjectByName
	FindStreamProcessorByName = findStreamProcessorByName
)
