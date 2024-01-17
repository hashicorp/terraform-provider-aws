// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package account

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/account"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindRegionOptInStatus(ctx context.Context, conn *account.Client, accountID, region string) (*account.GetRegionOptStatusOutput, error) {
	input := &account.GetRegionOptStatusInput{
		RegionName: aws.String(region),
	}
	if accountID != "" {
		input.AccountId = aws.String(accountID)
	}

	output, err := conn.GetRegionOptStatus(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
