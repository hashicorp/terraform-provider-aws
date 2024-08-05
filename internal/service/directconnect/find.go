// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindLagByID(ctx context.Context, conn *directconnect.Client, id string) (*awstypes.Lag, error) {
	input := &directconnect.DescribeLagsInput{
		LagId: aws.String(id),
	}

	output, err := conn.DescribeLags(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.DirectConnectClientException](err, "Could not find Lag with ID") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	lag, err := tfresource.AssertSingleValueResult(output.Lags)

	if err != nil {
		return nil, err
	}

	if lag.LagState == awstypes.LagStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(lag.LagState),
			LastRequest: input,
		}
	}

	return lag, nil
}

func FindLocationByCode(ctx context.Context, conn *directconnect.Client, code string) (awstypes.Location, error) {
	input := &directconnect.DescribeLocationsInput{}

	locations, err := FindLocations(ctx, conn, input)

	if err != nil {
		return awstypes.Location{}, err
	}

	for _, location := range locations {
		if aws.ToString(location.LocationCode) == code {
			return location, nil
		}
	}

	return awstypes.Location{}, tfresource.NewEmptyResultError(input)
}

func FindLocations(ctx context.Context, conn *directconnect.Client, input *directconnect.DescribeLocationsInput) ([]awstypes.Location, error) {
	output, err := conn.DescribeLocations(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Locations) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Locations, nil
}
