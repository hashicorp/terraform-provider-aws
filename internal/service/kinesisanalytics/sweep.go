//go:build sweep
// +build sweep

package kinesisanalytics

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_kinesis_analytics_application", &resource.Sweeper{
		Name: "aws_kinesis_analytics_application",
		F:    sweepApplications,
	})
}

func sweepApplications(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).KinesisAnalyticsConn
	input := &kinesisanalytics.ListApplicationsInput{}
	var sweeperErrs *multierror.Error

	err = ListApplicationsPages(conn, input, func(page *kinesisanalytics.ListApplicationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, applicationSummary := range page.ApplicationSummaries {
			arn := aws.StringValue(applicationSummary.ApplicationARN)
			name := aws.StringValue(applicationSummary.ApplicationName)

			application, err := FindApplicationDetailByName(conn, name)

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
			err = r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
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

	return sweeperErrs.ErrorOrNil()
}
