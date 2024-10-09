// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package deploy

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/codedeploy"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_codedeploy_app", &resource.Sweeper{
		Name: "aws_codedeploy_app",
		F:    sweepApps,
	})
}

func sweepApps(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.DeployClient(ctx)
	input := &codedeploy.ListApplicationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := codedeploy.NewListApplicationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CodeDeploy Application sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing CodeDeploy Applications (%s): %w", region, err)
		}

		for _, v := range page.Applications {
			r := resourceApp()
			d := r.Data(nil)
			d.SetId(fmt.Sprintf("%s:%s", "xxxx", v))
			d.Set(names.AttrName, v)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CodeDeploy Applications (%s): %w", region, err)
	}

	return nil
}
