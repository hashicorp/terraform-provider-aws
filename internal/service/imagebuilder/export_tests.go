// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

// Exports for use in tests only.
var (
	ResourceComponent       = resourceComponent
	ResourceContainerRecipe = resourceContainerRecipe
	ResourceLifecyclePolicy = newResourceLifecyclePolicy

	FindComponentByARN       = findComponentByARN
	FindContainerRecipeByARN = findContainerRecipeByARN
	FindLifecyclePolicyByARN = findLifecyclePolicyByARN
)
