//go:build sweep
// +build sweep

package ds

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).DSConn()

	sweepResources := make([]sweep.Sweepable, 0)

	input := &directoryservice.DescribeDirectoriesInput{}
	err = describeDirectoriesPages(ctx, conn, input, func(page *directoryservice.DescribeDirectoriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, directory := range page.DirectoryDescriptions {
			r := ResourceDirectory()
			d := r.Data(nil)
			d.SetId(aws.StringValue(directory.DirectoryId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Directory Service Directory sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing Directory Service Directories (%s): %w", region, err)
	}

	err = sweep.SweepOrchestratorWithContext(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping Directory Service Directories (%s): %w", region, err)
	}

	return nil
}

func sweepRegions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).DSConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &directoryservice.DescribeDirectoriesInput{}

	err = describeDirectoriesPages(ctx, conn, input, func(page *directoryservice.DescribeDirectoriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, directory := range page.DirectoryDescriptions {
			if directory == nil {
				continue
			}

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
					if region != nil && aws.StringValue(region.RegionType) != directoryservice.RegionTypePrimary {
						r := ResourceRegion()
						d := r.Data(nil)
						d.SetId(RegionCreateResourceID(aws.StringValue(region.DirectoryId), aws.StringValue(region.RegionName)))
						sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
					}
				}

				return !lastPage
			})

			if tfawserr.ErrMessageContains(err, directoryservice.ErrCodeUnsupportedOperationException, "Multi-region replication") {
				log.Printf("[INFO] Skipping Directory Service Regions for %s", aws.StringValue(directory.DirectoryId))
				continue
			}
			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("describing Directory Service Regions for %s: %w", aws.StringValue(directory.DirectoryId), err))
				continue
			}
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("listing Directory Service Directories for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("sweeping Directory Service Regions for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Directory Service Regions sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
