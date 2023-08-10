// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudhsmv2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindClusterByID(ctx context.Context, conn *cloudhsmv2.CloudHSMV2, id string) (*cloudhsmv2.Cluster, error) {
	input := &cloudhsmv2.DescribeClustersInput{
		Filters: map[string][]*string{
			"clusterIds": aws.StringSlice([]string{id}),
		},
	}

	output, err := findCluster(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == cloudhsmv2.ClusterStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.ClusterId) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findCluster(ctx context.Context, conn *cloudhsmv2.CloudHSMV2, input *cloudhsmv2.DescribeClustersInput) (*cloudhsmv2.Cluster, error) {
	output, err := findClusters(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func findClusters(ctx context.Context, conn *cloudhsmv2.CloudHSMV2, input *cloudhsmv2.DescribeClustersInput) ([]*cloudhsmv2.Cluster, error) {
	var output []*cloudhsmv2.Cluster

	err := conn.DescribeClustersPagesWithContext(ctx, input, func(page *cloudhsmv2.DescribeClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Clusters {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindHSMByTwoPartKey(ctx context.Context, conn *cloudhsmv2.CloudHSMV2, hsmID, eniID string) (*cloudhsmv2.Hsm, error) {
	input := &cloudhsmv2.DescribeClustersInput{}

	output, err := findClusters(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	for _, v := range output {
		for _, v := range v.Hsms {
			if v == nil {
				continue
			}

			// CloudHSMv2 HSM instances can be recreated, but the ENI ID will
			// remain consistent. Without this ENI matching, HSM instances
			// instances can become orphaned.
			if aws.StringValue(v.HsmId) == hsmID || aws.StringValue(v.EniId) == eniID {
				return v, nil
			}
		}
	}

	return nil, &retry.NotFoundError{}
}
