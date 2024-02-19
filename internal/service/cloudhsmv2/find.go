// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudhsmv2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudhsmv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

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
