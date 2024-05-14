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

func FindHostedZoneDNSSEC(ctx context.Context, conn *route53.Route53, hostedZoneID string) (*route53.GetDNSSECOutput, error) {
	input := &route53.GetDNSSECInput{
		HostedZoneId: aws.String(hostedZoneID),
	}

	output, err := conn.GetDNSSECWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindQueryLoggingConfigByID(ctx context.Context, conn *route53.Route53, id string) (*route53.QueryLoggingConfig, error) {
	input := &route53.GetQueryLoggingConfigInput{
		Id: aws.String(id),
	}

	output, err := conn.GetQueryLoggingConfigWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchQueryLoggingConfig, route53.ErrCodeNoSuchHostedZone) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.QueryLoggingConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.QueryLoggingConfig, nil
}

func FindTrafficPolicyByID(ctx context.Context, conn *route53.Route53, id string) (*route53.TrafficPolicy, error) {
	var latestVersion int64

	err := listTrafficPoliciesPages(ctx, conn, &route53.ListTrafficPoliciesInput{}, func(page *route53.ListTrafficPoliciesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TrafficPolicySummaries {
			if aws.StringValue(v.Id) == id {
				latestVersion = aws.Int64Value(v.LatestVersion)

				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if latestVersion == 0 {
		return nil, tfresource.NewEmptyResultError(id)
	}

	input := &route53.GetTrafficPolicyInput{
		Id:      aws.String(id),
		Version: aws.Int64(latestVersion),
	}

	output, err := conn.GetTrafficPolicyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchTrafficPolicy) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.TrafficPolicy == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.TrafficPolicy, nil
}

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
