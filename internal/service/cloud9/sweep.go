// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloud9

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/cloud9"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_cloud9_environment_ec2", &resource.Sweeper{
		Name: "aws_cloud9_environment_ec2",
		F:    sweepEnvironmentEC2s,
	})
}

func sweepEnvironmentEC2s(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.Cloud9Client(ctx)
	input := &cloud9.ListEnvironmentsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := cloud9.NewListEnvironmentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Cloud9 EC2 Environment sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Cloud9 EC2 Environments (%s): %w", region, err)
		}

		for _, v := range page.EnvironmentIds {
			r := resourceEnvironmentEC2()
			d := r.Data(nil)
			d.SetId(v)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Cloud9 EC2 Environments (%s): %w", region, err)
	}

	return nil
}
