// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoverycontrolconfig

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	r53rcc "github.com/aws/aws-sdk-go-v2/service/route53recoverycontrolconfig"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_route53recoverycontrolconfig_cluster", &resource.Sweeper{
		Name: "aws_route53recoverycontrolconfig_cluster",
		F:    sweepClusters,
		Dependencies: []string{
			"aws_route53recoverycontrolconfig_control_panel",
		},
	})

	resource.AddTestSweepers("aws_route53recoverycontrolconfig_control_panel", &resource.Sweeper{
		Name: "aws_route53recoverycontrolconfig_control_panel",
		F:    sweepControlPanels,
		Dependencies: []string{
			"aws_route53recoverycontrolconfig_routing_control",
			"aws_route53recoverycontrolconfig_safety_rule",
		},
	})

	resource.AddTestSweepers("aws_route53recoverycontrolconfig_routing_control", &resource.Sweeper{
		Name: "aws_route53recoverycontrolconfig_routing_control",
		F:    sweepRoutingControls,
	})

	resource.AddTestSweepers("aws_route53recoverycontrolconfig_safety_rule", &resource.Sweeper{
		Name: "aws_route53recoverycontrolconfig_safety_rule",
		F:    sweepSafetyRules,
	})
}

func sweepClusters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.Route53RecoveryControlConfigClient(ctx)
	input := &r53rcc.ListClustersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := r53rcc.NewListClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Route53 Recovery Control Config Cluster sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Route53 Recovery Control Config Clusters (%s): %w", region, err)
		}

		for _, v := range page.Clusters {
			r := resourceCluster()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ClusterArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Route53 Recovery Control Config Clusters (%s): %w", region, err)
	}

	return nil
}

func sweepControlPanels(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.Route53RecoveryControlConfigClient(ctx)
	input := &r53rcc.ListClustersInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := r53rcc.NewListClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Route53 Recovery Control Config Control Panel sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Route53 Recovery Control Config Clusters (%s): %w", region, err))
		}

		for _, v := range page.Clusters {
			input := &r53rcc.ListControlPanelsInput{
				ClusterArn: v.ClusterArn,
			}

			pages := r53rcc.NewListControlPanelsPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					continue
				}

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Route53 Recovery Control Config Control Panels (%s): %w", region, err))
				}

				for _, v := range page.ControlPanels {
					if aws.ToBool(v.DefaultControlPanel) {
						continue
					}

					r := resourceControlPanel()
					d := r.Data(nil)
					d.SetId(aws.ToString(v.ControlPanelArn))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Route53 Recovery Control Config Control Panels (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepRoutingControls(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.Route53RecoveryControlConfigClient(ctx)
	input := &r53rcc.ListClustersInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := r53rcc.NewListClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Route53 Recovery Control Config Routing Control sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Route53 Recovery Control Config Clusters (%s): %w", region, err))
		}

		for _, v := range page.Clusters {
			input := &r53rcc.ListControlPanelsInput{
				ClusterArn: v.ClusterArn,
			}

			pages := r53rcc.NewListControlPanelsPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					continue
				}

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Route53 Recovery Control Config Control Panels (%s): %w", region, err))
				}

				for _, v := range page.ControlPanels {
					input := &r53rcc.ListRoutingControlsInput{
						ControlPanelArn: v.ControlPanelArn,
					}

					pages := r53rcc.NewListRoutingControlsPaginator(conn, input)
					for pages.HasMorePages() {
						page, err := pages.NextPage(ctx)

						if awsv2.SkipSweepError(err) {
							continue
						}

						if err != nil {
							sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Route53 Recovery Control Config Routing Controls (%s): %w", region, err))
						}

						for _, v := range page.RoutingControls {
							r := resourceRoutingControl()
							d := r.Data(nil)
							d.SetId(aws.ToString(v.RoutingControlArn))

							sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
						}
					}
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Route53 Recovery Control Config Routing Controls (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepSafetyRules(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.Route53RecoveryControlConfigClient(ctx)
	input := &r53rcc.ListClustersInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := r53rcc.NewListClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Route53 Recovery Control Config Safety Rule sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Route53 Recovery Control Config Clusters (%s): %w", region, err))
		}

		for _, v := range page.Clusters {
			input := &r53rcc.ListControlPanelsInput{
				ClusterArn: v.ClusterArn,
			}

			pages := r53rcc.NewListControlPanelsPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					continue
				}

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Route53 Recovery Control Config Control Panels (%s): %w", region, err))
				}

				for _, v := range page.ControlPanels {
					input := &r53rcc.ListSafetyRulesInput{
						ControlPanelArn: v.ControlPanelArn,
					}

					pages := r53rcc.NewListSafetyRulesPaginator(conn, input)
					for pages.HasMorePages() {
						page, err := pages.NextPage(ctx)

						if awsv2.SkipSweepError(err) {
							continue
						}

						if err != nil {
							sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Route53 Recovery Control Config Safety Rules (%s): %w", region, err))
						}

						for _, v := range page.SafetyRules {
							r := resourceSafetyRule()
							d := r.Data(nil)
							if v.ASSERTION != nil {
								d.SetId(aws.ToString(v.ASSERTION.SafetyRuleArn))
							} else if v.GATING != nil {
								d.SetId(aws.ToString(v.GATING.SafetyRuleArn))
							} else {
								continue
							}

							sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
						}
					}
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Route53 Recovery Control Config Safety Rules (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}
