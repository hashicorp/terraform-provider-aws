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

	resource.AddTestSweepers("aws_codebuild_project", &resource.Sweeper{
		Name: "aws_codebuild_project",
		F:    sweepProjects,
	})

	resource.AddTestSweepers("aws_codebuild_source_credential", &resource.Sweeper{
		Name: "aws_codebuild_source_credential",
		F:    sweepSourceCredentials,
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

func sweepProjects(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).CodeBuildConn
	input := &codebuild.ListProjectsInput{}
	var sweeperErrs *multierror.Error

	err = conn.ListProjectsPages(input, func(page *codebuild.ListProjectsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, arn := range page.Projects {
			id := aws.StringValue(arn)
			r := ResourceProject()
			d := r.Data(nil)
			d.SetId(id)

			err := r.Delete(d, client)
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting CodeBuild Project (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CodeBuild Project sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving CodeBuild Projects: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepSourceCredentials(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).CodeBuildConn
	input := &codebuild.ListSourceCredentialsInput{}
	var sweeperErrs *multierror.Error

	creds, err := conn.ListSourceCredentials(input)

	for _, cred := range creds.SourceCredentialsInfos {
		id := aws.StringValue(cred.Arn)
		r := ResourceSourceCredential()
		d := r.Data(nil)
		d.SetId(id)

		err := r.Delete(d, client)
		if err != nil {
			sweeperErr := fmt.Errorf("error deleting CodeBuild Source Credential (%s): %w", id, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CodeBuild Source Credential sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving CodeBuild Source Credentials: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
