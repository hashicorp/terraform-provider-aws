// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apprunner"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_apprunner_auto_scaling_configuration_version", &resource.Sweeper{
		Name: "aws_apprunner_auto_scaling_configuration_version",
		F:    sweepAutoScalingConfigurationVersions,
		Dependencies: []string{
			"aws_apprunner_service",
		},
	})

	resource.AddTestSweepers("aws_apprunner_connection", &resource.Sweeper{
		Name: "aws_apprunner_connection",
		F:    sweepConnections,
		Dependencies: []string{
			"aws_apprunner_service",
		},
	})

	resource.AddTestSweepers("aws_apprunner_service", &resource.Sweeper{
		Name: "aws_apprunner_service",
		F:    sweepServices,
	})
}

func sweepAutoScalingConfigurationVersions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &apprunner.ListAutoScalingConfigurationsInput{}
	conn := client.AppRunnerClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := apprunner.NewListAutoScalingConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		output, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping App Runner AutoScaling Configuration sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing App Runner AutoScaling Configurations (%s): %w", region, err)
		}

		for _, v := range output.AutoScalingConfigurationSummaryList {
			arn := aws.ToString(v.AutoScalingConfigurationArn)

			// Skip DefaultConfigurations as deletion not supported by the AppRunner service
			// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/19840
			if aws.ToBool(v.IsDefault) {
				log.Printf("[INFO] Skipping App Runner Default AutoScaling Configuration: %s", arn)
				continue
			}

			r := resourceAutoScalingConfigurationVersion()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping App Runner AutoScaling Configurations (%s): %w", region, err)
	}

	return nil
}

func sweepConnections(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &apprunner.ListConnectionsInput{}
	conn := client.AppRunnerClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := apprunner.NewListConnectionsPaginator(conn, input)
	for pages.HasMorePages() {
		output, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping App Runner Connection sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing App Runner Connections (%s): %w", region, err)
		}

		for _, v := range output.ConnectionSummaryList {
			r := resourceConnection()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ConnectionName))
			d.Set(names.AttrARN, v.ConnectionArn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping App Runner Connections (%s): %w", region, err)
	}

	return nil
}

func sweepServices(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &apprunner.ListServicesInput{}
	conn := client.AppRunnerClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := apprunner.NewListServicesPaginator(conn, input)
	for pages.HasMorePages() {
		output, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping App Runner Service sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing App Runner Services (%s): %w", region, err)
		}

		for _, v := range output.ServiceSummaryList {
			r := resourceService()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ServiceArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping App Runner Services (%s): %w", region, err)
	}

	return nil
}
