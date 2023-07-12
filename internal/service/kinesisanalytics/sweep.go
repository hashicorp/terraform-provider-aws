// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package kinesisanalytics

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	conn := client.KinesisAnalyticsConn(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	input := &kinesisanalytics.ListApplicationsInput{}
	err = ListApplicationsPages(ctx, conn, input, func(page *kinesisanalytics.ListApplicationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, applicationSummary := range page.ApplicationSummaries {
			arn := aws.StringValue(applicationSummary.ApplicationARN)
			name := aws.StringValue(applicationSummary.ApplicationName)

			application, err := FindApplicationDetailByName(ctx, conn, name)

			if tfawserr.ErrMessageContains(err, kinesisanalytics.ErrCodeUnsupportedOperationException, "was created/updated by kinesisanalyticsv2 SDK") {
				continue
			}

			if err != nil {
				sweeperErr := fmt.Errorf("error reading Kinesis Analytics Application (%s): %w", arn, err)
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			r := ResourceApplication()
			d := r.Data(nil)
			d.SetId(arn)
			d.Set("create_timestamp", aws.TimeValue(application.CreateTimestamp).Format(time.RFC3339))
			d.Set("name", name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Kinesis Analytics Application sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Kinesis Analytics Applications: %w", err))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Kinesis Analytics Applications: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
