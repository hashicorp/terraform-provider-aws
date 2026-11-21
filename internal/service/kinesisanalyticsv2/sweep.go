// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kinesisanalyticsv2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_awstypes.application", &resource.Sweeper{
		Name: "aws_awstypes.application",
		F:    sweepApplication,
	})
	// Enhanced snapshots outlive their application and keep accruing storage charges.
	awsv2.Register("aws_kinesisanalyticsv2_snapshot", sweepSnapshots)
}

func sweepSnapshots(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.KinesisAnalyticsV2Client(ctx)
	var sweepResources []sweep.Sweepable

	// Unfiltered, so that imported snapshots, which have no source application, are included.
	input := kinesisanalyticsv2.ListSnapshotsInput{}
	pages := kinesisanalyticsv2.NewListSnapshotsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.SnapshotInfoList {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newSnapshotResource, client,
				sweepfw.NewAttribute(names.AttrName, aws.ToString(v.SnapshotName))))
		}
	}

	return sweepResources, nil
}

func sweepApplication(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.KinesisAnalyticsV2Client(ctx)
	input := &kinesisanalyticsv2.ListApplicationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := kinesisanalyticsv2.NewListApplicationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Kinesis Analytics v2 Application sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Kinesis Analytics v2 Applications (%s): %w", region, err)
		}

		for _, v := range page.ApplicationSummaries {
			arn := aws.ToString(v.ApplicationARN)
			name := aws.ToString(v.ApplicationName)

			application, err := findApplicationDetailByName(ctx, conn, name)

			if err != nil {
				continue
			}

			r := resourceApplication()
			d := r.Data(nil)
			d.SetId(arn)
			d.Set("create_timestamp", aws.ToTime(application.CreateTimestamp).Format(time.RFC3339))
			d.Set(names.AttrName, name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Kinesis Analytics v2 Applications (%s): %w", region, err)
	}

	return nil
}
