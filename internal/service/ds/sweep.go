//go:build sweep
// +build sweep

package ds

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
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
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).DSConn

	var sweeperErrs *multierror.Error

	input := &directoryservice.DescribeDirectoriesInput{}

	err = describeDirectoriesPagesWithContext(context.TODO(), conn, input, func(page *directoryservice.DescribeDirectoriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, directory := range page.DirectoryDescriptions {
			id := aws.StringValue(directory.DirectoryId)

			r := ResourceDirectory()
			d := r.Data(nil)
			d.SetId(id)

			err := r.Delete(d, client)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Directory Service Directory (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Directory Service Directory sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErr := fmt.Errorf("error listing Directory Service Directories: %w", err)
		log.Printf("[ERROR] %s", sweeperErr)
		sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepRegions(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).DSConn
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &directoryservice.DescribeDirectoriesInput{}

	err = describeDirectoriesPagesWithContext(context.TODO(), conn, input, func(page *directoryservice.DescribeDirectoriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, directory := range page.DirectoryDescriptions {
			if directory == nil {
				continue
			}

			id := aws.StringValue(directory.DirectoryId)

			err := describeRegionsPages(conn, &directoryservice.DescribeRegionsInput{
				DirectoryId: aws.String(id),
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

			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("describing DS Regions for %s: %w", region, err))
				continue
			}
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("describing DS Directories for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping DS Regions for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping DS Regions sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
