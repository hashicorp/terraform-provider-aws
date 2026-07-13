// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3files"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_s3files_file_system", sweepFileSystems, "aws_s3files_access_point", "aws_s3files_mount_target")
	awsv2.Register("aws_s3files_access_point", sweepAccessPoints)
	awsv2.Register("aws_s3files_mount_target", sweepMountTargets)
}

func sweepAccessPoints(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.S3FilesClient(ctx)
	var sweepResources []sweep.Sweepable

	input := s3files.ListFileSystemsInput{}
	output, err := conn.ListFileSystems(ctx, &input)
	if err != nil {
		return nil, err
	}

	for _, fs := range output.FileSystems {
		apInput := s3files.ListAccessPointsInput{
			FileSystemId: fs.FileSystemId,
		}
		apOutput, err := conn.ListAccessPoints(ctx, &apInput)
		if err != nil {
			continue
		}

		for _, ap := range apOutput.AccessPoints {
			sweepResources = append(sweepResources, framework.NewSweepResource(newAccessPointResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(ap.AccessPointId)),
			))
		}
	}

	return sweepResources, nil
}

func sweepFileSystems(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.S3FilesClient(ctx)
	var sweepResources []sweep.Sweepable

	input := s3files.ListFileSystemsInput{}
	output, err := conn.ListFileSystems(ctx, &input)
	if err != nil {
		return nil, err
	}

	for _, v := range output.FileSystems {
		sweepResources = append(sweepResources, framework.NewSweepResource(newFileSystemResource, client,
			framework.NewAttribute(names.AttrID, aws.ToString(v.FileSystemId)),
		))
	}

	return sweepResources, nil
}

func sweepMountTargets(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.S3FilesClient(ctx)
	var sweepResources []sweep.Sweepable

	input := s3files.ListFileSystemsInput{}
	output, err := conn.ListFileSystems(ctx, &input)
	if err != nil {
		return nil, err
	}

	for _, fs := range output.FileSystems {
		mtInput := s3files.ListMountTargetsInput{
			FileSystemId: fs.FileSystemId,
		}
		mtOutput, err := conn.ListMountTargets(ctx, &mtInput)
		if err != nil {
			continue
		}

		for _, mt := range mtOutput.MountTargets {
			sweepResources = append(sweepResources, framework.NewSweepResource(newMountTargetResource, client,
				framework.NewAttribute(names.AttrID, aws.ToString(mt.MountTargetId)),
			))
		}
	}

	return sweepResources, nil
}
