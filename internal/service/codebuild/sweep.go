//go:build sweep
// +build sweep

package codebuild

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_codebuild_report_group", &resource.Sweeper{
		Name: "aws_codebuild_report_group",
		F:    sweepReportGroups,
	})
}

func sweepReportGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).CodeBuildConn
	input := &codebuild.ListReportGroupsInput{}
	var sweeperErrs *multierror.Error

	err = conn.ListReportGroupsPages(input, func(page *codebuild.ListReportGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, arn := range page.ReportGroups {
			id := aws.StringValue(arn)
			r := ResourceReportGroup()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("delete_reports", true)

			err := r.Delete(d, client)
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting CodeBuild Report Group (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CodeBuild Report Group sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving CodeBuild ReportGroups: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
