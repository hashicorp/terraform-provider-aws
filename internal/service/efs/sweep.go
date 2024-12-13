// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_efs_access_point", &resource.Sweeper{
		Name: "aws_efs_access_point",
		F:    sweepAccessPoints,
	})

	resource.AddTestSweepers("aws_efs_file_system", &resource.Sweeper{
		Name: "aws_efs_file_system",
		F:    sweepFileSystems,
		Dependencies: []string{
			"aws_efs_mount_target",
			"aws_efs_access_point",
			"aws_m2_environment",
		},
	})

	resource.AddTestSweepers("aws_efs_mount_target", &resource.Sweeper{
		Name: "aws_efs_mount_target",
		F:    sweepMountTargets,
	})
}

func sweepAccessPoints(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.EFSClient(ctx)
	input := &efs.DescribeFileSystemsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := efs.NewDescribeFileSystemsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EFS Access Point sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EFS File Systems (%s): %w", region, err)
		}

		for _, v := range page.FileSystems {
			input := &efs.DescribeAccessPointsInput{
				FileSystemId: v.FileSystemId,
			}

			pages := efs.NewDescribeAccessPointsPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					continue
				}

				for _, v := range page.AccessPoints {
					r := resourceAccessPoint()
					d := r.Data(nil)
					d.SetId(aws.ToString(v.AccessPointId))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EFS Acess Points (%s): %w", region, err)
	}

	return nil
}

func sweepFileSystems(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.EFSClient(ctx)
	input := &efs.DescribeFileSystemsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := efs.NewDescribeFileSystemsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EFS File System sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EFS File Systems (%s): %w", region, err)
		}

		for _, v := range page.FileSystems {
			r := resourceFileSystem()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.FileSystemId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EFS File Systems (%s): %w", region, err)
	}

	return nil
}

func sweepMountTargets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.EFSClient(ctx)
	input := &efs.DescribeFileSystemsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := efs.NewDescribeFileSystemsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EFS Mount Target sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing EFS File Systems (%s): %w", region, err)
		}

		for _, v := range page.FileSystems {
			input := &efs.DescribeMountTargetsInput{
				FileSystemId: v.FileSystemId,
			}

			pages := efs.NewDescribeMountTargetsPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					continue
				}

				for _, v := range page.MountTargets {
					r := resourceMountTarget()
					d := r.Data(nil)
					d.SetId(aws.ToString(v.MountTargetId))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping EFS Mount Targets (%s): %w", region, err)
	}

	return nil
}
