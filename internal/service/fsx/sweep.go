// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_fsx_backup", &resource.Sweeper{
		Name: "aws_fsx_backup",
		F:    sweepBackups,
	})

	resource.AddTestSweepers("aws_fsx_lustre_file_system", &resource.Sweeper{
		Name: "aws_fsx_lustre_file_system",
		F:    sweepLustreFileSystems,
		Dependencies: []string{
			"aws_datasync_location",
			"aws_m2_environment",
		},
	})

	resource.AddTestSweepers("aws_fsx_ontap_file_system", &resource.Sweeper{
		Name: "aws_fsx_ontap_file_system",
		F:    sweepONTAPFileSystems,
		Dependencies: []string{
			"aws_datasync_location",
			"aws_fsx_ontap_storage_virtual_machine",
			"aws_m2_environment",
		},
	})

	resource.AddTestSweepers("aws_fsx_ontap_storage_virtual_machine", &resource.Sweeper{
		Name: "aws_fsx_ontap_storage_virtual_machine",
		F:    sweepONTAPStorageVirtualMachine,
		Dependencies: []string{
			"aws_fsx_ontap_volume",
		},
	})

	resource.AddTestSweepers("aws_fsx_ontap_volume", &resource.Sweeper{
		Name: "aws_fsx_ontap_volume",
		F:    sweepONTAPVolumes,
	})

	resource.AddTestSweepers("aws_fsx_openzfs_file_system", &resource.Sweeper{
		Name: "aws_fsx_openzfs_file_system",
		F:    sweepOpenZFSFileSystems,
		Dependencies: []string{
			"aws_datasync_location",
			"aws_fsx_openzfs_volume",
			"aws_m2_environment",
		},
	})

	resource.AddTestSweepers("aws_fsx_openzfs_volume", &resource.Sweeper{
		Name: "aws_fsx_openzfs_volume",
		F:    sweepOpenZFSVolume,
	})

	resource.AddTestSweepers("aws_fsx_windows_file_system", &resource.Sweeper{
		Name: "aws_fsx_windows_file_system",
		F:    sweepWindowsFileSystems,
		Dependencies: []string{
			"aws_datasync_location",
			"aws_m2_environment",
			"aws_storagegateway_file_system_association",
		},
	})
}

func sweepBackups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.FSxClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	input := &fsx.DescribeBackupsInput{}

	pages := fsx.NewDescribeBackupsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error listing FSx Backups for %s: %w", region, err))
		}

		if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error sweeping FSx Backups for %s: %w", region, err))
		}

		for _, fs := range page.Backups {
			r := resourceBackup()
			d := r.Data(nil)
			d.SetId(aws.ToString(fs.BackupId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping FSx Backups sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepLustreFileSystems(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.FSxClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	input := &fsx.DescribeFileSystemsInput{}

	pages := fsx.NewDescribeFileSystemsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error listing FSx Lustre File Systems for %s: %w", region, err))
		}

		if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error sweeping FSx Lustre File Systems for %s: %w", region, err))
		}

		for _, fs := range page.FileSystems {
			if fs.FileSystemType != awstypes.FileSystemTypeLustre {
				continue
			}

			r := resourceLustreFileSystem()
			d := r.Data(nil)
			d.SetId(aws.ToString(fs.FileSystemId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping FSx Lustre File System sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepONTAPFileSystems(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.FSxClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	input := &fsx.DescribeFileSystemsInput{}

	pages := fsx.NewDescribeFileSystemsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error listing FSx ONTAP File Systems for %s: %w", region, err))
		}

		if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error sweeping FSx ONTAP File Systems for %s: %w", region, err))
		}

		for _, fs := range page.FileSystems {
			if fs.FileSystemType != awstypes.FileSystemTypeOntap {
				continue
			}

			r := resourceONTAPFileSystem()
			d := r.Data(nil)
			d.SetId(aws.ToString(fs.FileSystemId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping FSx ONTAP File System sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepONTAPStorageVirtualMachine(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.FSxClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	input := &fsx.DescribeStorageVirtualMachinesInput{}

	pages := fsx.NewDescribeStorageVirtualMachinesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error listing FSx ONTAP Storage Virtual Machine for %s: %w", region, err))
		}

		if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error sweeping FSx ONTAP Storage Virtual Machine for %s: %w", region, err))
		}

		for _, vm := range page.StorageVirtualMachines {
			r := resourceONTAPStorageVirtualMachine()
			d := r.Data(nil)
			d.SetId(aws.ToString(vm.StorageVirtualMachineId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping FSx ONTAP Storage Virtual Machine sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepONTAPVolumes(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.FSxClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	input := &fsx.DescribeVolumesInput{}

	pages := fsx.NewDescribeVolumesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error listing FSx ONTAP Volume for %s: %w", region, err))
		}

		if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error sweeping FSx ONTAP Volume for %s: %w", region, err))
		}

		for _, v := range page.Volumes {
			if v.VolumeType != awstypes.VolumeTypeOntap {
				continue
			}
			if v.OntapConfiguration != nil && aws.ToBool(v.OntapConfiguration.StorageVirtualMachineRoot) {
				continue
			}

			r := resourceONTAPVolume()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.VolumeId))
			d.Set("bypass_snaplock_enterprise_retention", true)
			d.Set("skip_final_backup", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping FSx ONTAP Volume sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepOpenZFSFileSystems(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.FSxClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	input := &fsx.DescribeFileSystemsInput{}

	pages := fsx.NewDescribeFileSystemsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error listing FSx OpenZFS File Systems for %s: %w", region, err))
		}

		if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error sweeping FSx OpenZFS File Systems for %s: %w", region, err))
		}

		for _, fs := range page.FileSystems {
			if fs.FileSystemType != awstypes.FileSystemTypeOpenzfs {
				continue
			}

			r := resourceOpenZFSFileSystem()
			d := r.Data(nil)
			d.SetId(aws.ToString(fs.FileSystemId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping FSx OpenZFS File System sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepOpenZFSVolume(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.FSxClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	input := &fsx.DescribeVolumesInput{}

	pages := fsx.NewDescribeVolumesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error listing FSx OpenZFS Volume for %s: %w", region, err))
		}

		if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error sweeping FSx OpenZFS Volume for %s: %w", region, err))
		}

		for _, v := range page.Volumes {
			if v.VolumeType != awstypes.VolumeTypeOpenzfs {
				continue
			}
			if v.OpenZFSConfiguration != nil && aws.ToString(v.OpenZFSConfiguration.ParentVolumeId) == "" {
				continue
			}

			r := resourceOpenZFSVolume()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.VolumeId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping FSx OpenZFS Volume sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepWindowsFileSystems(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.FSxClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	input := &fsx.DescribeFileSystemsInput{}

	pages := fsx.NewDescribeFileSystemsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error listing FSx Windows File Systems for %s: %w", region, err))
		}

		if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("error sweeping FSx Windows File Systems for %s: %w", region, err))
		}

		for _, fs := range page.FileSystems {
			if fs.FileSystemType != awstypes.FileSystemTypeWindows {
				continue
			}

			r := resourceWindowsFileSystem()
			d := r.Data(nil)
			d.SetId(aws.ToString(fs.FileSystemId))
			d.Set("skip_final_backup", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if awsv2.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping FSx Windows File System sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
