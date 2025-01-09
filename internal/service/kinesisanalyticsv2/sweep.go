// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesisanalyticsv2

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_awstypes.application", &resource.Sweeper{
		Name: "aws_awstypes.application",
		F:    sweepApplication,
	})
}

func sweepApplication(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
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
