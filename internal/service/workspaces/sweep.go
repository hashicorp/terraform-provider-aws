//go:build sweep
// +build sweep

package workspaces

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_workspaces_directory", &resource.Sweeper{
		Name:         "aws_workspaces_directory",
		F:            sweepDirectories,
		Dependencies: []string{"aws_workspaces_workspace", "aws_workspaces_ip_group"},
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
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).WorkSpacesConn()
	input := &workspaces.DescribeWorkspaceDirectoriesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeWorkspaceDirectoriesPagesWithContext(ctx, input, func(page *workspaces.DescribeWorkspaceDirectoriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, directory := range page.Directories {
			r := ResourceDirectory()
			d := r.Data(nil)
			d.SetId(aws.StringValue(directory.DirectoryId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WorkSpaces Directory sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WorkSpaces Directories (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WorkSpaces Directories (%s): %w", region, err)
	}

	return nil
}

func sweepIPGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).WorkSpacesConn()
	input := &workspaces.DescribeIpGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = describeIPGroupsPages(ctx, conn, input, func(page *workspaces.DescribeIpGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, ipGroup := range page.Result {
			r := ResourceIPGroup()
			d := r.Data(nil)
			d.SetId(aws.StringValue(ipGroup.GroupId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping WorkSpaces Ip Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing WorkSpaces Ip Groups (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping WorkSpaces Ip Groups (%s): %w", region, err)
	}

	return nil
}

func sweepWorkspace(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).WorkSpacesConn()

	var errors error
	input := &workspaces.DescribeWorkspacesInput{}
	err = conn.DescribeWorkspacesPagesWithContext(ctx, input, func(resp *workspaces.DescribeWorkspacesOutput, _ bool) bool {
		for _, workspace := range resp.Workspaces {
			err := WorkspaceDelete(ctx, conn, aws.StringValue(workspace.WorkspaceId), WorkspaceTerminatedTimeout)
			if err != nil {
				errors = multierror.Append(errors, err)
			}

		}
		return true
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping workspaces sweep for %s: %s", region, err)
		return errors // In case we have completed some pages, but had errors
	}
	if err != nil {
		errors = multierror.Append(errors, fmt.Errorf("error listing workspaces: %s", err))
	}

	return errors
}
