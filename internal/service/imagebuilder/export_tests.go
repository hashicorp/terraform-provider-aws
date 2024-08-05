// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

// Exports for use in tests only.
var (
	ResourceComponent                 = resourceComponent
	ResourceContainerRecipe           = resourceContainerRecipe
	ResourceDistributionConfiguration = resourceDistributionConfiguration
	ResourceImage                     = resourceImage
	ResourceImagePipeline             = resourceImagePipeline
	ResourceImageRecipe               = resourceImageRecipe
	ResourceLifecyclePolicy           = newResourceLifecyclePolicy

	FindComponentByARN                 = findComponentByARN
	FindContainerRecipeByARN           = findContainerRecipeByARN
	FindDistributionConfigurationByARN = findDistributionConfigurationByARN
	FindImageByARN                     = findImageByARN
	FindImagePipelineByARN             = findImagePipelineByARN
	FindImageRecipeByARN               = findImageRecipeByARN
	FindLifecyclePolicyByARN           = findLifecyclePolicyByARN
)
