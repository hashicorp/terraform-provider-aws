// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusHostedZoneDNSSEC(ctx context.Context, conn *route53.Route53, hostedZoneID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		hostedZoneDnssec, err := FindHostedZoneDNSSEC(ctx, conn, hostedZoneID)

		if err != nil {
			return nil, "", err
		}

		if hostedZoneDnssec == nil || hostedZoneDnssec.Status == nil {
			return nil, "", nil
		}

		return hostedZoneDnssec.Status, aws.StringValue(hostedZoneDnssec.Status.ServeSignature), nil
	}
}

func statusTrafficPolicyInstanceState(ctx context.Context, conn *route53.Route53, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTrafficPolicyInstanceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}
