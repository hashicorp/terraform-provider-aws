// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package codebuild

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.CodeBuildConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	input := &codebuild.ListReportGroupsInput{}
	err = conn.ListReportGroupsPagesWithContext(ctx, input, func(page *codebuild.ListReportGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, arn := range page.ReportGroups {
			id := aws.StringValue(arn)
			r := ResourceReportGroup()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("delete_reports", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CodeBuild Report Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving CodeBuild ReportGroups: %w", err)
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping CodeBuild ReportGroups: %w", err)
	}

	return nil
}

func sweepProjects(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.CodeBuildConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	input := &codebuild.ListProjectsInput{}
	err = conn.ListProjectsPagesWithContext(ctx, input, func(page *codebuild.ListProjectsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, arn := range page.Projects {
			id := aws.StringValue(arn)
			r := ResourceProject()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CodeBuild Project sweep for %s: %s", region, err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("error retrieving CodeBuild Projects: %w", err)
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping CodeBuild Projects: %w", err)
	}

	return nil
}

func sweepSourceCredentials(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.CodeBuildConn(ctx)
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	input := &codebuild.ListSourceCredentialsInput{}
	creds, err := conn.ListSourceCredentialsWithContext(ctx, input)

	for _, cred := range creds.SourceCredentialsInfos {
		id := aws.StringValue(cred.Arn)
		r := ResourceSourceCredential()
		d := r.Data(nil)
		d.SetId(id)

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CodeBuild Source Credential sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving CodeBuild Source Credentials: %w", err))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping CodeBuild Source Credentials: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
