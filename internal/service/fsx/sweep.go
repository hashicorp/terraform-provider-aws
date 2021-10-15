//go:build sweep
// +build sweep

package fsx

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_fsx_backup", &resource.Sweeper{
		Name: "aws_fsx_backup",
		F:    sweepFSXBackups,
	})

	resource.AddTestSweepers("aws_fsx_lustre_file_system", &resource.Sweeper{
		Name: "aws_fsx_lustre_file_system",
		F:    sweepFSXLustreFileSystems,
	})

	resource.AddTestSweepers("aws_fsx_ontap_file_system", &resource.Sweeper{
		Name: "aws_fsx_ontap_file_system",
		F:    sweepFSXOntapFileSystems,
	})

	resource.AddTestSweepers("aws_fsx_windows_file_system", &resource.Sweeper{
		Name: "aws_fsx_windows_file_system",
		F:    sweepFSXWindowsFileSystems,
	})
}

func sweepFSXBackups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).FSxConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error
	input := &fsx.DescribeBackupsInput{}

	err = conn.DescribeBackupsPages(input, func(page *fsx.DescribeBackupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, fs := range page.Backups {
			r := ResourceBackup()
			d := r.Data(nil)
			d.SetId(aws.StringValue(fs.BackupId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing FSx Backups for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping FSx Backups for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping FSx Backups sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepFSXLustreFileSystems(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).FSxConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error
	input := &fsx.DescribeFileSystemsInput{}

	err = conn.DescribeFileSystemsPages(input, func(page *fsx.DescribeFileSystemsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, fs := range page.FileSystems {
			if aws.StringValue(fs.FileSystemType) != fsx.FileSystemTypeLustre {
				continue
			}

			r := ResourceLustreFileSystem()
			d := r.Data(nil)
			d.SetId(aws.StringValue(fs.FileSystemId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing FSx Lustre File Systems for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping FSx Lustre File Systems for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping FSx Lustre File System sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepFSXOntapFileSystems(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).FSxConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error
	input := &fsx.DescribeFileSystemsInput{}

	err = conn.DescribeFileSystemsPages(input, func(page *fsx.DescribeFileSystemsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, fs := range page.FileSystems {
			if aws.StringValue(fs.FileSystemType) != fsx.FileSystemTypeOntap {
				continue
			}

			r := ResourceOntapFileSystem()
			d := r.Data(nil)
			d.SetId(aws.StringValue(fs.FileSystemId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing FSx Ontap File Systems for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping FSx Ontap File Systems for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping FSx Ontap File System sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepFSXWindowsFileSystems(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).FSxConn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error
	input := &fsx.DescribeFileSystemsInput{}

	err = conn.DescribeFileSystemsPages(input, func(page *fsx.DescribeFileSystemsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, fs := range page.FileSystems {
			if aws.StringValue(fs.FileSystemType) != fsx.FileSystemTypeWindows {
				continue
			}

			r := ResourceWindowsFileSystem()
			d := r.Data(nil)
			d.SetId(aws.StringValue(fs.FileSystemId))
			d.Set("skip_final_backup", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing FSx Windows File Systems for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestrator(sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping FSx Windows File Systems for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping FSx Windows File System sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
