// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudtrail

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
	tforganizations "github.com/hashicorp/terraform-provider-aws/internal/service/organizations"
)

const (
	cloudTrailServicePrincipal = "cloudtrail.amazonaws.com"
)

func FindDelegatedAccountByAccountID(ctx context.Context, conn *organizations.Client, accountID string) (*types.DelegatedAdministrator, error) {

	account, err := tforganizations.FindDelegatedAdministratorByTwoPartKey(ctx, conn, accountID, cloudTrailServicePrincipal)
	if err != nil {
		return nil, err
	}

	return account, nil
}
