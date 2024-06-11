// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	tferrs "github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_directory_service_directory", &resource.Sweeper{
		Name: "aws_directory_service_directory",
		F:    sweepDirectories,
		Dependencies: []string{
			"aws_appstream_directory_config",
			"aws_connect_instance",
			"aws_db_instance",
			"aws_ec2_client_vpn_endpoint",
			"aws_fsx_ontap_storage_virtual_machine",
			"aws_fsx_windows_file_system",
			"aws_transfer_server",
			"aws_workspaces_directory",
			"aws_directory_service_region",
		},
	})

	resource.AddTestSweepers("aws_directory_service_region", &resource.Sweeper{
		Name: "aws_directory_service_region",
		F:    sweepRegions,
	})
}

func sweepDirectories(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.DSClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)

	input := &directoryservice.DescribeDirectoriesInput{}
	err = describeDirectoriesPages(ctx, conn, input, func(page *directoryservice.DescribeDirectoriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, directory := range page.DirectoryDescriptions {
			r := ResourceDirectory()
			d := r.Data(nil)
			d.SetId(aws.ToString(directory.DirectoryId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv2.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Directory Service Directory sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing Directory Service Directories (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping Directory Service Directories (%s): %w", region, err)
	}

	return nil
}

func sweepRegions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.DSClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &directoryservice.DescribeDirectoriesInput{}

	err = describeDirectoriesPages(ctx, conn, input, func(page *directoryservice.DescribeDirectoriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, directory := range page.DirectoryDescriptions {
			if directory.RegionsInfo == nil || len(directory.RegionsInfo.AdditionalRegions) == 0 {
				continue
			}

			err := describeRegionsPages(ctx, conn, &directoryservice.DescribeRegionsInput{
				DirectoryId: directory.DirectoryId,
			}, func(page *directoryservice.DescribeRegionsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, region := range page.RegionsDescription {
					if region.RegionType != awstypes.RegionTypePrimary {
						r := ResourceRegion()
						d := r.Data(nil)
						d.SetId(RegionCreateResourceID(aws.ToString(region.DirectoryId), aws.ToString(region.RegionName)))
						sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
					}
				}

				return !lastPage
			})

			if tferrs.IsAErrorMessageContains[*awstypes.UnsupportedOperationException](err, "Multi-region replication") {
				log.Printf("[INFO] Skipping Directory Service Regions for %s", aws.ToString(directory.DirectoryId))
				continue
			}
			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("describing Directory Service Regions for %s: %w", aws.ToString(directory.DirectoryId), err))
				continue
			}
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("listing Directory Service Directories for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Directory Service Regions for %s: %w", region, err))
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Directory Service Regions sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
