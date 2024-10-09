// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package mwaa

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/mwaa"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_mwaa_environment", &resource.Sweeper{
		Name: "aws_mwaa_environment",
		F:    sweepEnvironment,
	})
}

func sweepEnvironment(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &mwaa.ListEnvironmentsInput{}
	conn := client.MWAAClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := mwaa.NewListEnvironmentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) || tfawserr.ErrCodeEquals(err, "InternalFailure") {
			log.Printf("[WARN] Skipping MWAA Environment sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing MWAA Environments (%s): %w", region, err)
		}

		for _, v := range page.Environments {
			r := resourceEnvironment()
			d := r.Data(nil)
			d.SetId(v)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping MWAA Environments (%s): %w", region, err)
	}

	return nil
}
