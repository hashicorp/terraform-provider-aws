// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin

// Exports for use in tests only.
var (
	ResourceCentralizationRuleForOrganization  = newCentralizationRuleForOrganizationResource
	ResourceDatasetIntegration                 = newDatasetIntegrationResource
	ResourceS3TableIntegration                 = newS3TableIntegrationResource
	ResourceTelemetryEnrichment                = newTelemetryEnrichmentResource
	ResourceTelemetryEvaluation                = newTelemetryEvaluationResource
	ResourceTelemetryEvaluationForOrganization = newTelemetryEvaluationForOrganizationResource
	ResourceTelemetryPipeline                  = newTelemetryPipelineResource
	ResourceTelemetryRule                      = newTelemetryRuleResource
	ResourceTelemetryRuleForOrganization       = newTelemetryRuleForOrganizationResource

	FindCentralizationRuleForOrganizationByID = findCentralizationRuleForOrganizationByID
	FindDatasetIntegrationByARN               = findDatasetIntegrationByARN
	FindS3TableIntegrationByARN               = findS3TableIntegrationByARN
	FindTelemetryEnrichment                   = findTelemetryEnrichment
	FindTelemetryEvaluation                   = findTelemetryEvaluation
	FindTelemetryEvaluationForOrganization    = findTelemetryEvaluationForOrganization
	FindTelemetryPipelineByARN                = findTelemetryPipelineByARN
	FindTelemetryRuleByName                   = findTelemetryRuleByName
	FindTelemetryRuleForOrganizationByName    = findTelemetryRuleForOrganizationByName
)
