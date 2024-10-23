// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesisanalytics

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesisanalytics"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kinesisanalytics/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_kinesis_analytics_application", &resource.Sweeper{
		Name: "aws_kinesis_analytics_application",
		F:    sweepApplications,
	})
}

func sweepApplications(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.KinesisAnalyticsClient(ctx)
	input := &kinesisanalytics.ListApplicationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = listApplicationsPages(ctx, conn, input, func(page *kinesisanalytics.ListApplicationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ApplicationSummaries {
			arn := aws.ToString(v.ApplicationARN)
			name := aws.ToString(v.ApplicationName)

			application, err := findApplicationDetailByName(ctx, conn, name)

			if errs.IsAErrorMessageContains[*awstypes.UnsupportedOperationException](err, "was created/updated by kinesisanalyticsv2 SDK") {
				continue
			}

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

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Kinesis Analytics Application sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing Kinesis Analytics Applications (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Kinesis Analytics Applications (%s): %w", region, err)
	}

	return nil
}
