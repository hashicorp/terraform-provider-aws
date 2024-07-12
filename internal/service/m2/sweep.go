// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package m2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/m2"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	sweep.Register("aws_m2_application", sweepApplications)

	sweep.Register("aws_m2_environment", sweepEnvironments)
}

func sweepApplications(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.M2Client(ctx)

	var sweepResources []sweep.Sweepable

	pages := m2.NewListApplicationsPaginator(conn, &m2.ListApplicationsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			tflog.Warn(ctx, "Skipping sweeper", map[string]any{
				"error": err.Error(),
			})
			return nil, nil
		}
		if err != nil {
			return nil, err
		}

		for _, application := range page.Applications {
			sweepResources = append(sweepResources, framework.NewSweepResource(newApplicationResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(application.ApplicationId))))
		}
	}

	return sweepResources, nil
}

func sweepEnvironments(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.M2Client(ctx)

	var sweepResources []sweep.Sweepable

	pages := m2.NewListEnvironmentsPaginator(conn, &m2.ListEnvironmentsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			tflog.Warn(ctx, "Skipping sweeper", map[string]any{
				"error": err.Error(),
			})
			return nil, nil
		}
		if err != nil {
			return nil, err
		}

		for _, environment := range page.Environments {
			sweepResources = append(sweepResources, framework.NewSweepResource(newEnvironmentResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(environment.EnvironmentId))))
		}
	}

	return sweepResources, nil
}
