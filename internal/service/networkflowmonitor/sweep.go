// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkflowmonitor

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkflowmonitor"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
)

func RegisterSweepers() {
	awsv2.Register("aws_networkflowmonitor_monitor", sweepMonitors)
	awsv2.Register("aws_networkflowmonitor_scope", sweepScopes)
}

func sweepMonitors(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.NetworkFlowMonitorClient(ctx)
	var input networkflowmonitor.ListMonitorsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkflowmonitor.NewListMonitorsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Monitors {
			sweepResources = append(sweepResources, framework.NewSweepResource(newMonitorResource, client,
				framework.NewAttribute("monitor_name", aws.ToString(v.MonitorName)),
			))
		}
	}

	return sweepResources, nil
}

func sweepScopes(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.NetworkFlowMonitorClient(ctx)
	var input networkflowmonitor.ListScopesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := networkflowmonitor.NewListScopesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Scopes {
			sweepResources = append(sweepResources, framework.NewSweepResource(newScopeResource, client,
				framework.NewAttribute("scope_id", aws.ToString(v.ScopeId)),
			))
		}
	}

	return sweepResources, nil
}
