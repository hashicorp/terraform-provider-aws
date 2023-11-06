// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusCluster(ctx context.Context, conn *emr.EMR, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &emr.DescribeClusterInput{
			ClusterId: aws.String(id),
		}

		output, err := FindCluster(ctx, conn, input)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.State), nil
	}
}
