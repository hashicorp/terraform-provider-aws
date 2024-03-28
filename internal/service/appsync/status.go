// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/appsync"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func StatusAPICache(ctx context.Context, conn *appsync.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindAPICacheByID(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusDomainNameAPIAssociation(ctx context.Context, conn *appsync.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDomainNameAPIAssociationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.AssociationStatus), nil
	}
}
