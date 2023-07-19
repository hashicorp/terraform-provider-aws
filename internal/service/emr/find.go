// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindCluster(ctx context.Context, conn *emr.EMR, input *emr.DescribeClusterInput) (*emr.Cluster, error) {
	output, err := conn.DescribeClusterWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, ErrCodeClusterNotFound) || tfawserr.ErrMessageContains(err, emr.ErrCodeInvalidRequestException, "is not valid") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Cluster == nil || output.Cluster.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Cluster, nil
}

func FindClusterByID(ctx context.Context, conn *emr.EMR, id string) (*emr.Cluster, error) {
	input := &emr.DescribeClusterInput{
		ClusterId: aws.String(id),
	}

	output, err := FindCluster(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.Id) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	if state := aws.StringValue(output.Status.State); state == emr.ClusterStateTerminated || state == emr.ClusterStateTerminatedWithErrors {
		return nil, &retry.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return output, nil
}

func FindStudioByID(ctx context.Context, conn *emr.EMR, id string) (*emr.Studio, error) {
	input := &emr.DescribeStudioInput{
		StudioId: aws.String(id),
	}

	output, err := conn.DescribeStudioWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, emr.ErrCodeInvalidRequestException, "Studio does not exist") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Studio == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Studio, nil
}

func FindStudioSessionMappingByID(ctx context.Context, conn *emr.EMR, id string) (*emr.SessionMappingDetail, error) {
	studioId, identityType, identityId, err := readStudioSessionMapping(id)
	if err != nil {
		return nil, err
	}

	input := &emr.GetStudioSessionMappingInput{
		StudioId:     aws.String(studioId),
		IdentityType: aws.String(identityType),
		IdentityId:   aws.String(identityId),
	}

	output, err := conn.GetStudioSessionMappingWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, emr.ErrCodeInvalidRequestException, "Studio session mapping does not exist") ||
		tfawserr.ErrMessageContains(err, emr.ErrCodeInvalidRequestException, "Studio does not exist") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.SessionMapping == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.SessionMapping, nil
}

func FindBlockPublicAccessConfiguration(ctx context.Context, conn *emr.EMR) (*emr.GetBlockPublicAccessConfigurationOutput, error) {
	input := &emr.GetBlockPublicAccessConfigurationInput{}
	output, err := conn.GetBlockPublicAccessConfigurationWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if output == nil || output.BlockPublicAccessConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
