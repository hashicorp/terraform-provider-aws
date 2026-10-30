// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockdataautomation

// Exports for use in tests only.
var (
	ResourceBlueprint = newBlueprintResource
	ResourceProject   = newProjectResource

	FindBlueprintByARN = findBlueprintByARN
	FindProjectByARN   = findProjectByARN
)
