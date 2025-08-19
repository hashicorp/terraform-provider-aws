// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	awsv2.Register("aws_codebuild_report_group", sweepReportGroups)
	awsv2.Register("aws_codebuild_project", sweepProjects)
	awsv2.Register("aws_codebuild_source_credential", sweepSourceCredentials)
	awsv2.Register("aws_codebuild_fleet", sweepFleets)
}

func sweepReportGroups(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.CodeBuildClient(ctx)
	var input codebuild.ListReportGroupsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := codebuild.NewListReportGroupsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.ReportGroups {
			r := resourceReportGroup()
			d := r.Data(nil)
			d.SetId(v)
			d.Set("delete_reports", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepProjects(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.CodeBuildClient(ctx)
	var input codebuild.ListProjectsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := codebuild.NewListProjectsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Projects {
			r := resourceProject()
			d := r.Data(nil)
			d.SetId(v)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepSourceCredentials(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.CodeBuildClient(ctx)
	var input codebuild.ListSourceCredentialsInput
	sweepResources := make([]sweep.Sweepable, 0)

	output, err := conn.ListSourceCredentials(ctx, &input)

	if err != nil {
		return nil, err
	}

	for _, v := range output.SourceCredentialsInfos {
		id := aws.ToString(v.Arn)
		r := resourceSourceCredential()
		d := r.Data(nil)
		d.SetId(id)

		sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
	}

	return sweepResources, nil
}

func sweepFleets(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.CodeBuildClient(ctx)
	var input codebuild.ListFleetsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := codebuild.NewListFleetsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Fleets {
			r := resourceFleet()
			d := r.Data(nil)
			d.SetId(v)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
