// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindTrafficPolicyInstanceByID(ctx context.Context, conn *route53.Route53, id string) (*route53.TrafficPolicyInstance, error) {
	input := &route53.GetTrafficPolicyInstanceInput{
		Id: aws.String(id),
	}

	output, err := conn.GetTrafficPolicyInstanceWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchTrafficPolicyInstance) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.TrafficPolicyInstance == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.TrafficPolicyInstance, nil
}
