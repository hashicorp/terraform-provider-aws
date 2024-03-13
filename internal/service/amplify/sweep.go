// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amplify

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amplify"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_amplify_app", &resource.Sweeper{
		Name: "aws_amplify_app",
		F:    sweepApps,
	})
}

func sweepApps(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.AmplifyClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	input := &amplify.ListAppsInput{}
	err = listAppsPages(ctx, conn, input, func(page *amplify.ListAppsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, app := range page.Apps {
			r := ResourceApp()
			d := r.Data(nil)
			d.SetId(aws.ToString(app.AppId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Amplify App sweep for %s: %s", region, err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("error listing Amplify Apps: %w", err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Amplify Apps (%s): %w", region, err)
	}

	return nil
}
