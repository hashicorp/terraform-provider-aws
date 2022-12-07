//go:build sweep
// +build sweep

package efs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).EFSConn
	var sweeperErrs *multierror.Error

	var errors error
	input := &efs.DescribeFileSystemsInput{}
	err = conn.DescribeFileSystemsPages(input, func(page *efs.DescribeFileSystemsOutput, lastPage bool) bool {
		for _, filesystem := range page.FileSystems {
			id := aws.StringValue(filesystem.FileSystemId)
			log.Printf("[INFO] Deleting access points for EFS File System: %s", id)

			input := &efs.DescribeAccessPointsInput{
				FileSystemId: filesystem.FileSystemId,
			}
			for {
				out, err := conn.DescribeAccessPoints(input)
				if err != nil {
					errors = multierror.Append(errors, fmt.Errorf("error retrieving EFS access points on File System %q: %w", id, err))
					break
				}

				if out == nil || len(out.AccessPoints) == 0 {
					log.Printf("[INFO] No EFS access points to sweep on File System %q", id)
					break
				}

				for _, AccessPoint := range out.AccessPoints {
					id := aws.StringValue(AccessPoint.AccessPointId)

					log.Printf("[INFO] Deleting EFS access point: %s", id)
					r := ResourceAccessPoint()
					d := r.Data(nil)
					d.SetId(id)
					err := r.Delete(d, client)

					if err != nil {
						log.Printf("[ERROR] %s", err)
						sweeperErrs = multierror.Append(sweeperErrs, err)
						continue
					}
				}

				if out.NextToken == nil {
					break
				}
				input.NextToken = out.NextToken
			}
		}
		return true
	})
	if err != nil {
		errors = multierror.Append(errors, fmt.Errorf("error retrieving EFS File Systems: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepFileSystems(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EFSConn
	var sweeperErrs *multierror.Error

	input := &efs.DescribeFileSystemsInput{}
	err = conn.DescribeFileSystemsPages(input, func(page *efs.DescribeFileSystemsOutput, lastPage bool) bool {
		for _, filesystem := range page.FileSystems {
			id := aws.StringValue(filesystem.FileSystemId)

			log.Printf("[INFO] Deleting EFS File System: %s", id)

			r := ResourceFileSystem()
			d := r.Data(nil)
			d.SetId(id)
			err := r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}
		return true
	})
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving EFS File Systems: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepMountTargets(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).EFSConn

	var errors error
	input := &efs.DescribeFileSystemsInput{}
	err = conn.DescribeFileSystemsPages(input, func(page *efs.DescribeFileSystemsOutput, lastPage bool) bool {
		for _, filesystem := range page.FileSystems {
			id := aws.StringValue(filesystem.FileSystemId)
			log.Printf("[INFO] Deleting Mount Targets for EFS File System: %s", id)

			var errors error
			input := &efs.DescribeMountTargetsInput{
				FileSystemId: filesystem.FileSystemId,
			}
			for {
				out, err := conn.DescribeMountTargets(input)
				if err != nil {
					errors = multierror.Append(errors, fmt.Errorf("error retrieving EFS Mount Targets on File System %q: %w", id, err))
					break
				}

				if out == nil || len(out.MountTargets) == 0 {
					log.Printf("[INFO] No EFS Mount Targets to sweep on File System %q", id)
					break
				}

				for _, mounttarget := range out.MountTargets {
					id := aws.StringValue(mounttarget.MountTargetId)

					log.Printf("[INFO] Deleting EFS Mount Target: %s", id)
					_, err := conn.DeleteMountTarget(&efs.DeleteMountTargetInput{
						MountTargetId: mounttarget.MountTargetId,
					})
					if err != nil {
						errors = multierror.Append(errors, fmt.Errorf("error deleting EFS Mount Target %q: %w", id, err))
						continue
					}

					err = WaitForDeleteMountTarget(conn, id, mountTargetDeleteTimeout)
					if err != nil {
						errors = multierror.Append(errors, fmt.Errorf("error waiting for EFS Mount Target %q to delete: %w", id, err))
						continue
					}
				}

				if out.NextMarker == nil {
					break
				}
				input.Marker = out.NextMarker
			}
		}
		return true
	})
	if err != nil {
		errors = multierror.Append(errors, fmt.Errorf("error retrieving EFS File Systems: %w", err))
	}

	return errors
}
