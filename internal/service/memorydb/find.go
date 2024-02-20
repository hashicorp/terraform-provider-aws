// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/memorydb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindACLByName(ctx context.Context, conn *memorydb.MemoryDB, name string) (*memorydb.ACL, error) {
	input := memorydb.DescribeACLsInput{
		ACLName: aws.String(name),
	}

	output, err := conn.DescribeACLsWithContext(ctx, &input)

	if tfawserr.ErrCodeEquals(err, memorydb.ErrCodeACLNotFoundFault) {
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

func FindClusterByName(ctx context.Context, conn *memorydb.MemoryDB, name string) (*memorydb.Cluster, error) {
	input := memorydb.DescribeClustersInput{
		ClusterName:      aws.String(name),
		ShowShardDetails: aws.Bool(true),
	}

	output, err := conn.DescribeClustersWithContext(ctx, &input)

	if tfawserr.ErrCodeEquals(err, memorydb.ErrCodeClusterNotFoundFault) {
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

func FindParameterGroupByName(ctx context.Context, conn *memorydb.MemoryDB, name string) (*memorydb.ParameterGroup, error) {
	input := memorydb.DescribeParameterGroupsInput{
		ParameterGroupName: aws.String(name),
	}

	output, err := conn.DescribeParameterGroupsWithContext(ctx, &input)

	if tfawserr.ErrCodeEquals(err, memorydb.ErrCodeParameterGroupNotFoundFault) {
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

func FindSnapshotByName(ctx context.Context, conn *memorydb.MemoryDB, name string) (*memorydb.Snapshot, error) {
	input := memorydb.DescribeSnapshotsInput{
		SnapshotName: aws.String(name),
	}

	output, err := conn.DescribeSnapshotsWithContext(ctx, &input)

	if tfawserr.ErrCodeEquals(err, memorydb.ErrCodeSnapshotNotFoundFault) {
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

func FindSubnetGroupByName(ctx context.Context, conn *memorydb.MemoryDB, name string) (*memorydb.SubnetGroup, error) {
	input := memorydb.DescribeSubnetGroupsInput{
		SubnetGroupName: aws.String(name),
	}

	output, err := conn.DescribeSubnetGroupsWithContext(ctx, &input)

	if tfawserr.ErrCodeEquals(err, memorydb.ErrCodeSubnetGroupNotFoundFault) {
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

func FindUserByName(ctx context.Context, conn *memorydb.MemoryDB, name string) (*memorydb.User, error) {
	input := memorydb.DescribeUsersInput{
		UserName: aws.String(name),
	}

	output, err := conn.DescribeUsersWithContext(ctx, &input)

	if tfawserr.ErrCodeEquals(err, memorydb.ErrCodeUserNotFoundFault) {
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
