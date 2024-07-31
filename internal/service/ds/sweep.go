// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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
	input := &directoryservice.DescribeDirectoriesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := directoryservice.NewDescribeDirectoriesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Directory Service Directory sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Directory Service Directories (%s): %w", region, err)
		}

		for _, v := range page.DirectoryDescriptions {
			r := resourceDirectory()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.DirectoryId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Directory Service Directories (%s): %w", region, err)
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
	input := &directoryservice.DescribeDirectoriesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := directoryservice.NewDescribeDirectoriesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Directory Service Region sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Directory Service Directories (%s): %w", region, err)
		}

		for _, v := range page.DirectoryDescriptions {
			if v.RegionsInfo == nil || len(v.RegionsInfo.AdditionalRegions) == 0 {
				continue
			}

			input := &directoryservice.DescribeRegionsInput{
				DirectoryId: v.DirectoryId,
			}

			pages := directoryservice.NewDescribeRegionsPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					continue
				}

				for _, v := range page.RegionsDescription {
					if v.RegionType != awstypes.RegionTypePrimary {
						r := resourceRegion()
						d := r.Data(nil)
						d.SetId(regionCreateResourceID(aws.ToString(v.DirectoryId), aws.ToString(v.RegionName)))

						sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
					}
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Directory Service Regions (%s): %w", region, err)
	}

	return nil
}
