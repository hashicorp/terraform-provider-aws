// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func FindEndpointByID(ctx context.Context, conn *neptune.Neptune, id string) (*neptune.DBClusterEndpoint, error) {
	clusterId, endpointId, err := readClusterEndpointID(id)
	if err != nil {
		return nil, err
	}
	input := &neptune.DescribeDBClusterEndpointsInput{
		DBClusterIdentifier:         aws.String(clusterId),
		DBClusterEndpointIdentifier: aws.String(endpointId),
	}

	output, err := conn.DescribeDBClusterEndpointsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBClusterEndpointNotFoundFault) ||
		tfawserr.ErrCodeEquals(err, neptune.ErrCodeDBClusterNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &retry.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	endpoints := output.DBClusterEndpoints
	if len(endpoints) == 0 {
		return nil, &retry.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return endpoints[0], nil
}
