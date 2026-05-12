// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin

// Exports for use in tests only.
var (
	ResourceCentralizationRuleForOrganization  = newCentralizationRuleForOrganizationResource
	ResourceTelemetryEnrichment                = newTelemetryEnrichmentResource
	ResourceTelemetryEvaluation                = newTelemetryEvaluationResource
	ResourceTelemetryEvaluationForOrganization = newTelemetryEvaluationForOrganizationResource
	ResourceTelemetryPipeline                  = newTelemetryPipelineResource

	FindCentralizationRuleForOrganizationByID = findCentralizationRuleForOrganizationByID
	FindTelemetryEnrichment                   = findTelemetryEnrichment
	FindTelemetryEvaluation                   = findTelemetryEvaluation
	FindTelemetryEvaluationForOrganization    = findTelemetryEvaluationForOrganization
	FindTelemetryPipelineByARN                = findTelemetryPipelineByARN
)
