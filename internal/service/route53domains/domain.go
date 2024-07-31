// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53domains

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	"github.com/aws/aws-sdk-go-v2/service/route53domains/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findDomainDetailByName(ctx context.Context, conn *route53domains.Client, name string) (*route53domains.GetDomainDetailOutput, error) {
	input := &route53domains.GetDomainDetailInput{
		DomainName: aws.String(name),
	}

	output, err := conn.GetDomainDetail(ctx, input)

	if errs.IsAErrorMessageContains[*types.InvalidInput](err, "not found") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findDNSSECKeyByTwoPartKey(ctx context.Context, conn *route53domains.Client, domainName, keyID string) (*types.DnssecKey, error) {
	output, err := findDomainDetailByName(ctx, conn, domainName)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(tfslices.Filter(output.DnssecKeys, func(v types.DnssecKey) bool {
		return aws.ToString(v.Id) == keyID
	}))
}

func findDNSSECKeyByThreePartKey(ctx context.Context, conn *route53domains.Client, domainName string, flags int, publicKey string) (*types.DnssecKey, error) {
	output, err := findDomainDetailByName(ctx, conn, domainName)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(tfslices.Filter(output.DnssecKeys, func(v types.DnssecKey) bool {
		return int(aws.ToInt32(v.Flags)) == flags && aws.ToString(v.PublicKey) == publicKey
	}))
}
