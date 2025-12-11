// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_fsx_backup", sweepBackups)
	awsv2.Register("aws_fsx_lustre_file_system", sweepLustreFileSystems, "aws_datasync_location", "aws_m2_environment")
	awsv2.Register("aws_fsx_ontap_file_system", sweepONTAPFileSystems, "aws_datasync_location", "aws_fsx_ontap_storage_virtual_machine", "aws_m2_environment")
	awsv2.Register("aws_fsx_ontap_storage_virtual_machine", sweepONTAPStorageVirtualMachine, "aws_fsx_ontap_volume")
	awsv2.Register("aws_fsx_ontap_volume", sweepONTAPVolumes)
	awsv2.Register("aws_fsx_openzfs_file_system", sweepOpenZFSFileSystems, "aws_datasync_location", "aws_fsx_openzfs_volume", "aws_m2_environment")
	awsv2.Register("aws_fsx_openzfs_volume", sweepOpenZFSVolume, "aws_fsx_s3_access_point_attachment")
	awsv2.Register("aws_fsx_s3_access_point_attachment", sweepS3AccessPointAttachments)
	awsv2.Register("aws_fsx_windows_file_system", sweepWindowsFileSystems, "aws_datasync_location", "aws_m2_environment", "aws_storagegateway_file_system_association")
}

func sweepBackups(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.FSxClient(ctx)
	var input fsx.DescribeBackupsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := fsx.NewDescribeBackupsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Backups {
			r := resourceBackup()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.BackupId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepLustreFileSystems(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.FSxClient(ctx)
	var input fsx.DescribeFileSystemsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := fsx.NewDescribeFileSystemsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.FileSystems {
			if v.FileSystemType != awstypes.FileSystemTypeLustre {
				continue
			}

			r := resourceLustreFileSystem()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.FileSystemId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepONTAPFileSystems(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.FSxClient(ctx)
	var input fsx.DescribeFileSystemsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := fsx.NewDescribeFileSystemsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.FileSystems {
			if v.FileSystemType != awstypes.FileSystemTypeOntap {
				continue
			}

			r := resourceONTAPFileSystem()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.FileSystemId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepONTAPStorageVirtualMachine(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.FSxClient(ctx)
	var input fsx.DescribeStorageVirtualMachinesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := fsx.NewDescribeStorageVirtualMachinesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.StorageVirtualMachines {
			r := resourceONTAPStorageVirtualMachine()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.StorageVirtualMachineId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepONTAPVolumes(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.FSxClient(ctx)
	var input fsx.DescribeVolumesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := fsx.NewDescribeVolumesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Volumes {
			if v.VolumeType != awstypes.VolumeTypeOntap {
				continue
			}

			// Skip root volumes.
			if v.OntapConfiguration != nil && aws.ToBool(v.OntapConfiguration.StorageVirtualMachineRoot) {
				continue
			}

			var bypassSnaplock bool
			if v.OntapConfiguration != nil && v.OntapConfiguration.SnaplockConfiguration != nil {
				bypassSnaplock = true
			}
			r := resourceONTAPVolume()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.VolumeId))
			d.Set("bypass_snaplock_enterprise_retention", bypassSnaplock)
			d.Set("skip_final_backup", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepOpenZFSFileSystems(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.FSxClient(ctx)
	var input fsx.DescribeFileSystemsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := fsx.NewDescribeFileSystemsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.FileSystems {
			if v.FileSystemType != awstypes.FileSystemTypeOpenzfs {
				continue
			}

			r := resourceOpenZFSFileSystem()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.FileSystemId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepOpenZFSVolume(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.FSxClient(ctx)
	var input fsx.DescribeVolumesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := fsx.NewDescribeVolumesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
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

	return sweepResources, nil
}

func sweepS3AccessPointAttachments(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.FSxClient(ctx)
	var input fsx.DescribeS3AccessPointAttachmentsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := fsx.NewDescribeS3AccessPointAttachmentsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.S3AccessPointAttachments {
			sweepResources = append(sweepResources, framework.NewSweepResource(newS3AccessPointAttachmentResource, client,
				framework.NewAttribute(names.AttrName, aws.ToString(v.Name))))
		}
	}

	return sweepResources, nil
}

func sweepWindowsFileSystems(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.FSxClient(ctx)
	var input fsx.DescribeFileSystemsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := fsx.NewDescribeFileSystemsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.FileSystems {
			if v.FileSystemType != awstypes.FileSystemTypeWindows {
				continue
			}

			r := resourceWindowsFileSystem()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.FileSystemId))
			d.Set("skip_final_backup", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
