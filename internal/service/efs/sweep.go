// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package efs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	conn := client.EFSConn(ctx)
	input := &efs.DescribeFileSystemsInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeFileSystemsPagesWithContext(ctx, input, func(page *efs.DescribeFileSystemsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.FileSystems {
			input := &efs.DescribeAccessPointsInput{
				FileSystemId: v.FileSystemId,
			}

			err := conn.DescribeAccessPointsPagesWithContext(ctx, input, func(page *efs.DescribeAccessPointsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.AccessPoints {
					r := ResourceAccessPoint()
					d := r.Data(nil)
					d.SetId(aws.StringValue(v.AccessPointId))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EFS Access Points (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EFS Access Point sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EFS File Systems (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping EFS Access Points (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepFileSystems(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.EFSConn(ctx)
	input := &efs.DescribeFileSystemsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeFileSystemsPagesWithContext(ctx, input, func(page *efs.DescribeFileSystemsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.FileSystems {
			r := ResourceFileSystem()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.FileSystemId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EFS File System sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing EFS File Systems (%s): %w", region, err)
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
	conn := client.EFSConn(ctx)
	input := &efs.DescribeFileSystemsInput{}
	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeFileSystemsPagesWithContext(ctx, input, func(page *efs.DescribeFileSystemsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.FileSystems {
			input := &efs.DescribeMountTargetsInput{
				FileSystemId: v.FileSystemId,
			}

			err := describeMountTargetsPages(ctx, conn, input, func(page *efs.DescribeMountTargetsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.MountTargets {
					r := ResourceMountTarget()
					d := r.Data(nil)
					d.SetId(aws.StringValue(v.MountTargetId))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EFS Mount Targets (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EFS Mount Target sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EFS File Systems (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping EFS Mount Targets (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}
