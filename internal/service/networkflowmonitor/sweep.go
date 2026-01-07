// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkflowmonitor

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkflowmonitor"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_networkflowmonitor_monitor", &resource.Sweeper{
		Name: "aws_networkflowmonitor_monitor",
		F:    sweepMonitors,
	})

	resource.AddTestSweepers("aws_networkflowmonitor_scope", &resource.Sweeper{
		Name: "aws_networkflowmonitor_scope",
		F:    sweepScopes,
	})
}

func sweepMonitors(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.NetworkFlowMonitorClient(ctx)
	input := networkflowmonitor.ListMonitorsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkflowmonitor.NewListMonitorsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Network Flow Monitor Monitor sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Network Flow Monitor Monitors (%s): %w", region, err)
		}

		for _, v := range page.Monitors {
			arn := aws.ToString(v.MonitorArn)
			name := aws.ToString(v.MonitorName)

			sweepResources = append(sweepResources, framework.NewSweepResource(newMonitorResource, client,
				framework.NewAttribute(names.AttrID, arn),
				framework.NewAttribute("monitor_name", name),
			))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Network Flow Monitor Monitors (%s): %w", region, err)
	}

	return nil
}
func sweepScopes(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.NetworkFlowMonitorClient(ctx)
	input := networkflowmonitor.ListScopesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkflowmonitor.NewListScopesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Network Flow Monitor Scope sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Network Flow Monitor Scopes (%s): %w", region, err)
		}

		for _, v := range page.Scopes {
			scopeId := aws.ToString(v.ScopeId)
			scopeArn := aws.ToString(v.ScopeArn)

			sweepResources = append(sweepResources, framework.NewSweepResource(newScopeResource, client,
				framework.NewAttribute(names.AttrID, scopeArn),
				framework.NewAttribute("scope_id", scopeId),
			))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Network Flow Monitor Scopes (%s): %w", region, err)
	}

	return nil
}
