//go:build sweep
// +build sweep

package apprunner

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).AppRunnerConn()
	sweepResources := make([]sweep.Sweepable, 0)

	var errs *multierror.Error

	input := &apprunner.ListAutoScalingConfigurationsInput{}

	err = conn.ListAutoScalingConfigurationsPagesWithContext(ctx, input, func(page *apprunner.ListAutoScalingConfigurationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, summaryConfig := range page.AutoScalingConfigurationSummaryList {
			if summaryConfig == nil {
				continue
			}

			// Skip DefaultConfigurations as deletion not supported by the AppRunner service
			// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/19840
			if aws.StringValue(summaryConfig.AutoScalingConfigurationName) == "DefaultConfiguration" {
				log.Printf("[INFO] Skipping App Runner AutoScaling Configuration: DefaultConfiguration")
				continue
			}

			arn := aws.StringValue(summaryConfig.AutoScalingConfigurationArn)

			log.Printf("[INFO] Deleting App Runner AutoScaling Configuration Version (%s)", arn)
			r := ResourceAutoScalingConfigurationVersion()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping App Runner AutoScaling Configuration Versions sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing App Runner AutoScaling Configuration Versions: %w", err))
	}

	if err = sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping App Runner AutoScaling Configuration Version for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping App Runner AutoScaling Configuration Versions sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepConnections(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).AppRunnerConn()
	sweepResources := make([]sweep.Sweepable, 0)

	var errs *multierror.Error

	input := &apprunner.ListConnectionsInput{}

	err = conn.ListConnectionsPagesWithContext(ctx, input, func(page *apprunner.ListConnectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, c := range page.ConnectionSummaryList {
			if c == nil {
				continue
			}

			name := aws.StringValue(c.ConnectionName)

			log.Printf("[INFO] Deleting App Runner Connection: %s", name)

			r := ResourceConnection()
			d := r.Data(nil)
			d.SetId(name)
			d.Set("arn", c.ConnectionArn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping App Runner Connections sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing App Runner Connections: %w", err))
	}

	if err = sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping App Runner Connections for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping App Runner Connections sweep for %s: %s", region, err)
		return nil // In case we have completed some pages, but had errors
	}

	return errs.ErrorOrNil()
}

func sweepServices(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).AppRunnerConn()
	sweepResources := make([]sweep.Sweepable, 0)

	var errs *multierror.Error

	input := &apprunner.ListServicesInput{}

	err = conn.ListServicesPagesWithContext(ctx, input, func(page *apprunner.ListServicesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, service := range page.ServiceSummaryList {
			if service == nil {
				continue
			}

			arn := aws.StringValue(service.ServiceArn)

			log.Printf("[INFO] Deleting App Runner Service: %s", arn)

			r := ResourceService()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping App Runner Services sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing App Runner Services: %w", err))
	}

	if err = sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping App Runner Services for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping App Runner Services sweep for %s: %s", region, err)
		return nil // In case we have completed some pages, but had errors
	}

	return errs.ErrorOrNil()
}
