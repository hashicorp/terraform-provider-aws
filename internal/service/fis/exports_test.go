// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package fis

// Exports for use in tests only.
var (
	ResourceExperimentTemplate         = resourceExperimentTemplate
	ResourceTargetAccountConfiguration = newResourceTargetAccountConfiguration
	ResourceSafetyLeverState           = newSafetyLeverStateResource

	FindExperimentTemplateByID         = findExperimentTemplateByID
	FindTargetAccountConfigurationByID = findTargetAccountConfigurationByID
	FindSafetyLever                    = findSafetyLever
)
