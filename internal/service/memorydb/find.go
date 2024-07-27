// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/memorydb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/memorydb/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindACLByName(ctx context.Context, conn *memorydb.Client, name string) (*awstypes.ACL, error) {
	input := memorydb.DescribeACLsInput{
		ACLName: aws.String(name),
	}

	output, err := conn.DescribeACLs(ctx, &input)

	if tfawserr.ErrCodeEquals(err, awstypes.ErrCodeACLNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.ACLs) == 0 || output.ACLs[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.ACLs); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.ACLs[0], nil
}

func FindClusterByName(ctx context.Context, conn *memorydb.Client, name string) (*awstypes.Cluster, error) {
	input := memorydb.DescribeClustersInput{
		ClusterName:      aws.String(name),
		ShowShardDetails: aws.Bool(true),
	}

	output, err := conn.DescribeClusters(ctx, &input)

	if tfawserr.ErrCodeEquals(err, awstypes.ErrCodeClusterNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Clusters) == 0 || output.Clusters[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Clusters); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.Clusters[0], nil
}

func FindParameterGroupByName(ctx context.Context, conn *memorydb.Client, name string) (*awstypes.ParameterGroup, error) {
	input := memorydb.DescribeParameterGroupsInput{
		ParameterGroupName: aws.String(name),
	}

	output, err := conn.DescribeParameterGroups(ctx, &input)

	if tfawserr.ErrCodeEquals(err, awstypes.ErrCodeParameterGroupNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.ParameterGroups) == 0 || output.ParameterGroups[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.ParameterGroups); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.ParameterGroups[0], nil
}

func FindSnapshotByName(ctx context.Context, conn *memorydb.Client, name string) (*awstypes.Snapshot, error) {
	input := memorydb.DescribeSnapshotsInput{
		SnapshotName: aws.String(name),
	}

	output, err := conn.DescribeSnapshots(ctx, &input)

	if tfawserr.ErrCodeEquals(err, awstypes.ErrCodeSnapshotNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Snapshots) == 0 || output.Snapshots[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Snapshots); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.Snapshots[0], nil
}

func FindSubnetGroupByName(ctx context.Context, conn *memorydb.Client, name string) (*awstypes.SubnetGroup, error) {
	input := memorydb.DescribeSubnetGroupsInput{
		SubnetGroupName: aws.String(name),
	}

	output, err := conn.DescribeSubnetGroups(ctx, &input)

	if tfawserr.ErrCodeEquals(err, awstypes.ErrCodeSubnetGroupNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.SubnetGroups) == 0 || output.SubnetGroups[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.SubnetGroups); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.SubnetGroups[0], nil
}

func FindUserByName(ctx context.Context, conn *memorydb.Client, name string) (*awstypes.User, error) {
	input := memorydb.DescribeUsersInput{
		UserName: aws.String(name),
	}

	output, err := conn.DescribeUsers(ctx, &input)

	if tfawserr.ErrCodeEquals(err, awstypes.ErrCodeUserNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Users) == 0 || output.Users[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Users); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.Users[0], nil
}
