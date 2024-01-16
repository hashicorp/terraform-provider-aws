// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codepipeline

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_codepipeline", &resource.Sweeper{
		Name: "aws_codepipeline",
		F:    sweepPipelines,
	})
}

func sweepPipelines(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &codepipeline.ListPipelinesInput{}
	conn := client.CodePipelineClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := codepipeline.NewListPipelinesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Codepipeline Pipeline sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Codepipeline Pipelines (%s): %w", region, err)
		}

		for _, v := range page.Pipelines {
			r := resourcePipeline()
			d := r.Data(nil)

			d.SetId(aws.ToString(v.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Codepipeline Pipelines (%s): %w", region, err)
	}

	return nil
}
