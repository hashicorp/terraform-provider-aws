// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func StatusChangeSet(ctx context.Context, conn *cloudformation.Client, stackID, changeSetName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindChangeSetByStackIDAndChangeSetName(ctx, conn, stackID, changeSetName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func StatusStackSet(ctx context.Context, conn *cloudformation.Client, name, callAs string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindStackSetByName(ctx, conn, name, callAs)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func StatusStackSetOperation(ctx context.Context, conn *cloudformation.Client, stackSetName, operationID, callAs string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindStackSetOperationByStackSetNameAndOperationID(ctx, conn, stackSetName, operationID, callAs)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}
