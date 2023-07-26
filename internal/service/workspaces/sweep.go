// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package workspaces

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func init() {
	resource.AddTestSweepers("aws_workspaces_directory", &resource.Sweeper{
		Name: "aws_workspaces_directory",
		F:    sweepDirectories,
		Dependencies: []string{
			"aws_workspaces_workspace",
			"aws_workspaces_ip_group",
		},
	})

	resource.AddTestSweepers("aws_workspaces_ip_group", &resource.Sweeper{
		Name: "aws_workspaces_ip_group",
		F:    sweepIPGroups,
	})

	resource.AddTestSweepers("aws_workspaces_workspace", &resource.Sweeper{
		Name: "aws_workspaces_workspace",
		F:    sweepWorkspace,
	})
}

func sweepDirectories(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &workspaces.DescribeWorkspaceDirectoriesInput{}
	conn := client.WorkSpacesClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	paginator := workspaces.NewDescribeWorkspaceDirectoriesPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping WorkSpaces Directory sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing WorkSpaces Directories (%s): %w", region, err)
		}

		for _, v := range page.Directories {
			r := ResourceDirectory()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.DirectoryId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WorkSpaces Directories (%s): %w", region, err)
	}

	return nil
}

func sweepIPGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.WorkSpacesClient(ctx)
	input := &workspaces.DescribeIpGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeIPGroupsPages(ctx, conn, input, func(page *workspaces.DescribeIpGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Result {
			r := ResourceIPGroup()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.GroupId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WorkSpaces IP Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WorkSpaces IP Groups (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WorkSpaces IP Groups (%s): %w", region, err)
	}

	return nil
}

func describeIPGroupsPages(ctx context.Context, conn *workspaces.Client, input *workspaces.DescribeIpGroupsInput, fn func(*workspaces.DescribeIpGroupsOutput, bool) bool) error {
	for {
		output, err := conn.DescribeIpGroups(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.ToString(output.NextToken) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextToken = output.NextToken
	}
	return nil
}

func sweepWorkspace(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	input := &workspaces.DescribeWorkspacesInput{}
	conn := client.WorkSpacesClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	paginator := workspaces.NewDescribeWorkspacesPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping WorkSpaces Workspace sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing WorkSpaces Workspaces (%s): %w", region, err)
		}

		for _, v := range page.Workspaces {
			r := ResourceWorkspace()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.WorkspaceId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WorkSpaces Workspaces (%s): %w", region, err)
	}

	return nil
}
