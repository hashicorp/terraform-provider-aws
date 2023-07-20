// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package emrserverless

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emrserverless"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_emrserverless_application", &resource.Sweeper{
		Name: "aws_emrserverless_application",
		F:    sweepApplications,
	})
}

func sweepApplications(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.EMRServerlessConn(ctx)
	input := &emrserverless.ListApplicationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListApplicationsPagesWithContext(ctx, input, func(page *emrserverless.ListApplicationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Applications {
			if aws.StringValue(v.State) == emrserverless.ApplicationStateTerminated {
				continue
			}

			r := ResourceApplication()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EMR Serverless Application sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EMR Serverless Applications (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EMR Serverless Applications (%s): %w", region, err)
	}

	return nil
}
