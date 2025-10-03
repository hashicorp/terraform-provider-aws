// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/transfer"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_transfer_server", sweepServers)
	awsv2.Register("aws_transfer_web_app", sweepWebApps)
	awsv2.Register("aws_transfer_workflow", sweepWorkflows, "aws_transfer_server")
}

func sweepServers(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.TransferClient(ctx)
	var input transfer.ListServersInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := transfer.NewListServersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.Servers {
			r := resourceServer()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ServerId))
			d.Set(names.AttrForceDestroy, true) // In lieu of an aws_transfer_user sweeper.
			d.Set("identity_provider_type", v.IdentityProviderType)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepWebApps(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.TransferClient(ctx)
	var input transfer.ListWebAppsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := transfer.NewListWebAppsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.WebApps {
			sweepResources = append(sweepResources, framework.NewSweepResource(newWebAppResource, client,
				framework.NewAttribute("web_app_id", aws.ToString(v.WebAppId))),
			)
		}
	}

	return sweepResources, nil
}

func sweepWorkflows(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.TransferClient(ctx)
	var input transfer.ListWorkflowsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := transfer.NewListWorkflowsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.Workflows {
			r := resourceWorkflow()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.WorkflowId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
