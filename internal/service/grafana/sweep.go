// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/grafana"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_grafana_workspace", &resource.Sweeper{
		Name: "aws_grafana_workspace",
		F:    sweepWorkSpaces,
	})
}

func sweepWorkSpaces(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.GrafanaClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	input := &grafana.ListWorkspacesInput{}

	pages := grafana.NewListWorkspacesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Grafana Workspace sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Grafana Workspace for %s: %w", region, err)
		}

		for _, workspace := range page.Workspaces {
			id := aws.ToString(workspace.Id)
			log.Printf("[INFO] Deleting Grafana Workspace: %s", id)
			r := ResourceWorkspace()
			d := r.Data(nil)
			d.SetId(id)

			if err != nil {
				return fmt.Errorf("error reading Grafana Workspace %s: %w", id, err)
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)
	if err != nil {
		return fmt.Errorf("error sweeping Grafana Workspace: %w", err)
	}

	return nil
}
