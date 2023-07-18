// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package logs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_cloudwatch_log_group", &resource.Sweeper{
		Name: "aws_cloudwatch_log_group",
		F:    sweepGroups,
		Dependencies: []string{
			"aws_api_gateway_rest_api",
			"aws_cloudhsm_v2_cluster",
			"aws_cloudtrail",
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
		F:    sweeplogQueryDefinitions,
	})

	resource.AddTestSweepers("aws_cloudwatch_log_resource_policy", &resource.Sweeper{
		Name: "aws_cloudwatch_log_resource_policy",
		F:    sweepResourcePolicies,
	})
}

func sweepGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	input := &cloudwatchlogs.DescribeLogGroupsInput{}
	conn := client.LogsConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeLogGroupsPagesWithContext(ctx, input, func(page *cloudwatchlogs.DescribeLogGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LogGroups {
			r := resourceGroup()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.LogGroupName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudWatch Logs Log Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudWatch Logs Log Groups (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CloudWatch Logs Log Groups (%s): %w", region, err)
	}

	return nil
}

func sweeplogQueryDefinitions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	input := &cloudwatchlogs.DescribeQueryDefinitionsInput{}
	conn := client.LogsConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeQueryDefinitionsPages(ctx, conn, input, func(page *cloudwatchlogs.DescribeQueryDefinitionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.QueryDefinitions {
			r := resourceQueryDefinition()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.QueryDefinitionId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
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
	conn := client.LogsConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeResourcePoliciesPages(ctx, conn, input, func(page *cloudwatchlogs.DescribeResourcePoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ResourcePolicies {
			r := resourceResourcePolicy()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.PolicyName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
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
