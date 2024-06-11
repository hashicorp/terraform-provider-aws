// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusDirectoryStage(ctx context.Context, conn *directoryservice.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDirectoryByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Stage), nil
	}
}

func statusDirectoryShareStatus(ctx context.Context, conn *directoryservice.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDirectoryByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ShareStatus), nil
	}
}

func statusDomainController(ctx context.Context, conn *directoryservice.Client, directoryID, domainControllerID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDomainController(ctx, conn, directoryID, domainControllerID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusRadius(ctx context.Context, conn *directoryservice.Client, directoryID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDirectoryByID(ctx, conn, directoryID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.RadiusStatus), nil
	}
}

func statusRegion(ctx context.Context, conn *directoryservice.Client, directoryID, regionName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindRegion(ctx, conn, directoryID, regionName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusSharedDirectory(ctx context.Context, conn *directoryservice.Client, ownerDirectoryID, sharedDirectoryID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSharedDirectory(ctx, conn, ownerDirectoryID, sharedDirectoryID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ShareStatus), nil
	}
}
