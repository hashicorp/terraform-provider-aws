// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_cloudwatch_log_anomaly_detector", sweepAnomalyDetectors)

	awsv2.Register("aws_cloudwatch_log_delivery", sweepDeliveries)
	awsv2.Register("aws_cloudwatch_log_delivery_destination", sweepDeliveryDestinations, "aws_cloudwatch_log_delivery")
	awsv2.Register("aws_cloudwatch_log_delivery_source", sweepDeliverySources, "aws_cloudwatch_log_delivery")

	awsv2.Register("aws_cloudwatch_log_destination", sweepDestinations)

	resource.AddTestSweepers("aws_cloudwatch_log_group", &resource.Sweeper{
		Name: "aws_cloudwatch_log_group",
		F:    sweepGroups,
		Dependencies: []string{
			"aws_api_gateway_rest_api",
			"aws_cloudhsm_v2_cluster",
			"aws_cloudtrail",
			"aws_cloudwatch_log_anomaly_detector",
			"aws_datasync_task",
			"aws_db_instance",
			"aws_directory_service_directory",
			"aws_ec2_client_vpn_endpoint",
			"aws_eks_cluster",
			"aws_elasticsearch_domain",
			"aws_flow_log",
			"aws_glue_job",
			"aws_kinesis_analytics_application",
			"aws_kinesis_firehose_delivery_stream",
			"aws_lambda_function",
			"aws_mq_broker",
			"aws_msk_cluster",
			"aws_rds_cluster",
			"aws_route53_query_log",
			"aws_sagemaker_endpoint",
			"aws_storagegateway_gateway",
		},
	})

	resource.AddTestSweepers("aws_cloudwatch_query_definition", &resource.Sweeper{
		Name: "aws_cloudwatch_query_definition",
		F:    sweepQueryDefinitions,
	})

	resource.AddTestSweepers("aws_cloudwatch_log_resource_policy", &resource.Sweeper{
		Name: "aws_cloudwatch_log_resource_policy",
		F:    sweepResourcePolicies,
	})
}

func sweepAnomalyDetectors(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := &cloudwatchlogs.ListLogAnomalyDetectorsInput{}
	conn := client.LogsClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cloudwatchlogs.NewListLogAnomalyDetectorsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.AnomalyDetectors {
			sweepResources = append(sweepResources, framework.NewSweepResource(newAnomalyDetectorResource, client,
				framework.NewAttribute(names.AttrARN, aws.ToString(v.AnomalyDetectorArn))))
		}
	}

	return sweepResources, nil
}

func sweepDeliveries(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := &cloudwatchlogs.DescribeDeliveriesInput{}
	conn := client.LogsClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cloudwatchlogs.NewDescribeDeliveriesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Deliveries {
			sweepResources = append(sweepResources, framework.NewSweepResource(newDeliveryResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(v.Id))))
		}
	}

	return sweepResources, nil
}

func sweepDeliveryDestinations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := &cloudwatchlogs.DescribeDeliveryDestinationsInput{}
	conn := client.LogsClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cloudwatchlogs.NewDescribeDeliveryDestinationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.DeliveryDestinations {
			sweepResources = append(sweepResources, framework.NewSweepResource(newDeliveryDestinationResource, client,
				framework.NewAttribute(names.AttrName, aws.ToString(v.Name))))
		}
	}

	return sweepResources, nil
}

func sweepDeliverySources(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := &cloudwatchlogs.DescribeDeliverySourcesInput{}
	conn := client.LogsClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cloudwatchlogs.NewDescribeDeliverySourcesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.DeliverySources {
			sweepResources = append(sweepResources, framework.NewSweepResource(newDeliverySourceResource, client,
				framework.NewAttribute(names.AttrName, aws.ToString(v.Name))))
		}
	}

	return sweepResources, nil
}

func sweepDestinations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := &cloudwatchlogs.DescribeDestinationsInput{}
	conn := client.LogsClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cloudwatchlogs.NewDescribeDestinationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Destinations {
			r := resourceQueryDefinition()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.DestinationName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	input := &cloudwatchlogs.DescribeLogGroupsInput{}
	conn := client.LogsClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cloudwatchlogs.NewDescribeLogGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CloudWatch Logs Log Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing CloudWatch Logs Log Groups (%s): %w", region, err)
		}

		for _, v := range page.LogGroups {
			r := resourceGroup()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.LogGroupName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudWatch Logs Log Groups (%s): %w", region, err)
	}

	return nil
}

func sweepQueryDefinitions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	input := &cloudwatchlogs.DescribeQueryDefinitionsInput{}
	conn := client.LogsClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeQueryDefinitionsPages(ctx, conn, input, func(page *cloudwatchlogs.DescribeQueryDefinitionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.QueryDefinitions {
			r := resourceQueryDefinition()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.QueryDefinitionId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudWatch Logs Query Definition sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudWatch Logs Query Definitions (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudWatch Logs Query Definitions (%s): %w", region, err)
	}

	return nil
}

func sweepResourcePolicies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	input := &cloudwatchlogs.DescribeResourcePoliciesInput{}
	conn := client.LogsClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeResourcePoliciesPages(ctx, conn, input, func(page *cloudwatchlogs.DescribeResourcePoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ResourcePolicies {
			r := resourceResourcePolicy()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.PolicyName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudWatch Logs Resource Policy sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudWatch Logs Resource Policies (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudWatch Logs Resource Policies (%s): %w", region, err)
	}

	return nil
}
