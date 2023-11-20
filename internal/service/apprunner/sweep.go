// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apprunner"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_apprunner_auto_scaling_configuration_version", &resource.Sweeper{
		Name:         "aws_apprunner_auto_scaling_configuration_version",
		F:            sweepAutoScalingConfigurationVersions,
		Dependencies: []string{"aws_apprunner_service"},
	})

	resource.AddTestSweepers("aws_apprunner_connection", &resource.Sweeper{
		Name:         "aws_apprunner_connection",
		F:            sweepConnections,
		Dependencies: []string{"aws_apprunner_service"},
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

	conn := client.AppRunnerClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	var errs *multierror.Error

	input := &apprunner.ListAutoScalingConfigurationsInput{}

	paginator := apprunner.NewListAutoScalingConfigurationsPaginator(conn, input, func(o *apprunner.ListAutoScalingConfigurationsPaginatorOptions) {
		o.StopOnDuplicateToken = true
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)

		if err != nil {
			return err
		}
		for _, summaryConfig := range output.AutoScalingConfigurationSummaryList {
			// Skip DefaultConfigurations as deletion not supported by the AppRunner service
			// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/19840
			if aws.ToString(summaryConfig.AutoScalingConfigurationName) == "DefaultConfiguration" {
				log.Printf("[INFO] Skipping App Runner AutoScaling Configuration: DefaultConfiguration")
				continue
			}

			r := resourceAutoScalingConfigurationVersion()
			d := r.Data(nil)
			d.SetId(aws.ToString(summaryConfig.AutoScalingConfigurationArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping App Runner AutoScaling Configuration Versions sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing App Runner AutoScaling Configuration Versions: %w", err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping App Runner AutoScaling Configuration Version for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping App Runner AutoScaling Configuration Versions sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepConnections(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.AppRunnerClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	var errs *multierror.Error

	input := &apprunner.ListConnectionsInput{}

	paginator := apprunner.NewListConnectionsPaginator(conn, input, func(o *apprunner.ListConnectionsPaginatorOptions) {
		o.StopOnDuplicateToken = true
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)

		if err != nil {
			return err
		}

		for _, c := range output.ConnectionSummaryList {
			r := resourceConnection()
			d := r.Data(nil)
			d.SetId(aws.ToString(c.ConnectionName))
			d.Set("arn", c.ConnectionArn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping App Runner Connections sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing App Runner Connections: %w", err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping App Runner Connections for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping App Runner Connections sweep for %s: %s", region, err)
		return nil // In case we have completed some pages, but had errors
	}

	return errs.ErrorOrNil()
}

func sweepServices(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.AppRunnerClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	var errs *multierror.Error

	input := &apprunner.ListServicesInput{}

	paginator := apprunner.NewListServicesPaginator(conn, input, func(o *apprunner.ListServicesPaginatorOptions) {
		o.StopOnDuplicateToken = true
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)

		if err != nil {
			return err
		}

		for _, service := range output.ServiceSummaryList {
			arn := aws.ToString(service.ServiceArn)

			log.Printf("[INFO] Deleting App Runner Service: %s", arn)

			r := ResourceService()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping App Runner Services sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing App Runner Services: %w", err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping App Runner Services for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping App Runner Services sweep for %s: %s", region, err)
		return nil // In case we have completed some pages, but had errors
	}

	return errs.ErrorOrNil()
}
