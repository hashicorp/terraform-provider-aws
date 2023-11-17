// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package deploy

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/codedeploy"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
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
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &codedeploy.ListApplicationsInput{}

	paginator := codedeploy.NewListApplicationsPaginator(conn, input, func(o *codedeploy.ListApplicationsPaginatorOptions) {
		o.StopOnDuplicateToken = true
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)

		if err != nil {
			return err
		}

		for _, app := range output.Applications {
			if app == "" {
				continue
			}

			r := resourceApp()
			d := r.Data(nil)
			d.SetId(fmt.Sprintf("%s:%s", "xxxx", app))
			d.Set("name", app)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing CodeDeploy Applications for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping CodeDeploy Applications for %s: %w", region, err))
	}

	if awsv1.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping CodeDeploy Applications sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
