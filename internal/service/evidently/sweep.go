// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package evidently

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/evidently"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_evidently_project", &resource.Sweeper{
		Name: "aws_evidently_project",
		F:    sweepProjects,
	})
	resource.AddTestSweepers("aws_evidently_segment", &resource.Sweeper{
		Name: "aws_evidently_segment",
		F:    sweepSegments,
	})
}

func sweepProjects(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.EvidentlyClient(ctx)
	input := &evidently.ListProjectsInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := evidently.NewListProjectsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Evidently Project sweep for %s: %s", region, err)
			return nil
		}

		for _, project := range page.Projects {
			r := ResourceProject()
			d := r.Data(nil)
			d.SetId(aws.ToString(project.Name))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Evidently Projects for %s: %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepSegments(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.EvidentlyClient(ctx)
	input := &evidently.ListSegmentsInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	pages := evidently.NewListSegmentsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Evidently Segments sweep for %s: %s", region, err)
			return nil
		}

		for _, segment := range page.Segments {
			r := ResourceSegment()
			d := r.Data(nil)
			d.SetId(aws.ToString(segment.Arn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Evidently Segments for %s: %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}
