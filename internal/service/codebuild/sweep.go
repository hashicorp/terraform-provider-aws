// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
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
	conn := client.CodeBuildClient(ctx)
	input := &codebuild.ListReportGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := codebuild.NewListReportGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CodeBuild Report Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing CodeBuild ReportGroups (%s): %w", region, err)
		}

		for _, v := range page.ReportGroups {
			r := resourceReportGroup()
			d := r.Data(nil)
			d.SetId(v)
			d.Set("delete_reports", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CodeBuild ReportGroups (%s): %w", region, err)
	}

	return nil
}

func sweepProjects(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.CodeBuildClient(ctx)
	input := &codebuild.ListProjectsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := codebuild.NewListProjectsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping CodeBuild Project sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing CodeBuild Projects (%s): %w", region, err)
		}

		for _, v := range page.Projects {
			r := resourceProject()
			d := r.Data(nil)
			d.SetId(v)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CodeBuild Projects (%s): %w", region, err)
	}

	return nil
}

func sweepSourceCredentials(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.CodeBuildClient(ctx)
	input := &codebuild.ListSourceCredentialsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.ListSourceCredentials(ctx, input)

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping CodeBuild Source Credential sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CodeBuild Source Credentials (%s): %w", region, err)
	}

	for _, v := range output.SourceCredentialsInfos {
		id := aws.ToString(v.Arn)
		r := resourceSourceCredential()
		d := r.Data(nil)
		d.SetId(id)

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping CodeBuild Source Credentials (%s): %w", region, err)
	}

	return nil
}
