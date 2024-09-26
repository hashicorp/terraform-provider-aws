// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codegurureviewer

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codegurureviewer"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_codegurureviewer", &resource.Sweeper{
		Name: "aws_codegurureviewer",
		F:    sweepAssociations,
	})
}

func sweepAssociations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &codegurureviewer.ListRepositoryAssociationsInput{}
	conn := client.CodeGuruReviewerClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = listRepositoryAssociationsPages(ctx, conn, input, func(page *codegurureviewer.ListRepositoryAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.RepositoryAssociationSummaries {
			r := resourceRepositoryAssociation()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CodeGuru Reviewer Repository Association sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CodeGuru Reviewer Repository Associations (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CodeGuru Reviewer Repository Associations (%s): %w", region, err)
	}

	return nil
}
