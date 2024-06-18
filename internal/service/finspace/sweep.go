// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package finspace

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/finspace"
	"github.com/aws/aws-sdk-go-v2/service/finspace/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_finspace_kx_environment", &resource.Sweeper{
		Name: "aws_finspace_kx_environment",
		F:    sweepKxEnvironments,
	})
}

func sweepKxEnvironments(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.FinSpaceClient(ctx)
	input := &finspace.ListKxEnvironmentsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := finspace.NewListKxEnvironmentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping FinSpace Kx Environment sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing FinSpace Kx Environments (%s): %w", region, err)
		}

		for _, v := range page.Environments {
			id := aws.ToString(v.EnvironmentId)

			switch status := v.Status; status {
			case types.EnvironmentStatusDeleted, types.EnvironmentStatusDeleting, types.EnvironmentStatusCreating, types.EnvironmentStatusFailedDeletion:
				log.Printf("[INFO] Skipping FinSpace Kx Environment %s: Status=%s", id, status)
				continue
			}

			r := ResourceKxEnvironment()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping FinSpace Kx Environments (%s): %w", region, err)
	}

	return nil
}
