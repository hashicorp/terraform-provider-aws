// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package observabilityadmin

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/observabilityadmin"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_observabilityadmin_s3_table_integration", sweepS3TableIntegrations)
	awsv2.Register("aws_observabilityadmin_telemetry_enrichment", sweepTelemetryEnrichment, "aws_observabilityadmin_telemetry_rule")
	awsv2.Register("aws_observabilityadmin_telemetry_evaluation", sweepTelemetryEvaluation)
	awsv2.Register("aws_observabilityadmin_telemetry_pipeline", sweepTelemetryPipelines)
	awsv2.Register("aws_observabilityadmin_telemetry_rule", sweepTelemetryRules)
}

func sweepS3TableIntegrations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ObservabilityAdminClient(ctx)
	var input observabilityadmin.ListS3TableIntegrationsInput
	var sweepResources []sweep.Sweepable

	pages := observabilityadmin.NewListS3TableIntegrationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.IntegrationSummaries {
			sweepResources = append(sweepResources, framework.NewSweepResource(newS3TableIntegrationResource, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(v.Arn)),
			))
		}
	}

	return sweepResources, nil
}

func sweepTelemetryEnrichment(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var sweepResources []sweep.Sweepable

	// Singleton.
	sweepResources = append(sweepResources, framework.NewSweepResource(newTelemetryEnrichmentResource, client,
		framework.NewAttribute(names.AttrID, client.Region(ctx)),
	))

	return sweepResources, nil
}

func sweepTelemetryEvaluation(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	var sweepResources []sweep.Sweepable

	// Singleton.
	sweepResources = append(sweepResources, framework.NewSweepResource(newTelemetryEvaluationResource, client,
		framework.NewAttribute(names.AttrID, client.Region(ctx)),
	))

	return sweepResources, nil
}

func sweepTelemetryPipelines(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ObservabilityAdminClient(ctx)
	var input observabilityadmin.ListTelemetryPipelinesInput
	var sweepResources []sweep.Sweepable

	pages := observabilityadmin.NewListTelemetryPipelinesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.PipelineSummaries {
			sweepResources = append(sweepResources, framework.NewSweepResource(newTelemetryPipelineResource, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(v.Arn)),
			))
		}
	}

	return sweepResources, nil
}

func sweepTelemetryRules(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ObservabilityAdminClient(ctx)
	var input observabilityadmin.ListTelemetryRulesInput
	var sweepResources []sweep.Sweepable

	pages := observabilityadmin.NewListTelemetryRulesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.TelemetryRuleSummaries {
			sweepResources = append(sweepResources, framework.NewSweepResource(newTelemetryRuleResource, client,
				framework.NewAttribute("rule_name", aws.ToString(v.RuleName)),
			))
		}
	}

	return sweepResources, nil
}
