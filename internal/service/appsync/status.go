// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func StatusAPICache(ctx context.Context, conn *appsync.AppSync, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindAPICacheByID(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusDomainNameAPIAssociation(ctx context.Context, conn *appsync.AppSync, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDomainNameAPIAssociationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.AssociationStatus), nil
	}
}
