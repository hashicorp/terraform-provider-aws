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

func statusCluster(ctx context.Context, conn *cloudhsmv2.CloudHSMV2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClusterByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), err
	}
}

func statusHSM(ctx context.Context, conn *cloudhsmv2.CloudHSMV2, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindHSMByTwoPartKey(ctx, conn, id, "")

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), err
	}
}
