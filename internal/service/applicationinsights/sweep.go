// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package applicationinsights

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationinsights"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_applicationinsights_application", &resource.Sweeper{
		Name: "aws_applicationinsights_application",
		F:    sweepApplications,
	})
}

func sweepApplications(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.ApplicationInsightsClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	pages := applicationinsights.NewListApplicationsPaginator(conn, &applicationinsights.ListApplicationsInput{})

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error describing ApplicationInsights Applications: %w", err))
			// in case work can be done, don't jump out yet
		}

		for _, c := range page.ApplicationInfoList {
			r := ResourceApplication()
			d := r.Data(nil)
			d.SetId(aws.ToString(c.ResourceGroupName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping ApplicationInsights Applications for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping ApplicationInsights Application sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}
