// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/qbusiness"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func waitApplicationCreated(ctx context.Context, conn *qbusiness.QBusiness, id string, timeout time.Duration) (*qbusiness.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{qbusiness.ApplicationStatusCreating, qbusiness.ApplicationStatusUpdating},
		Target:     []string{qbusiness.ApplicationStatusActive},
		Refresh:    statusAppAvailability(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qbusiness.GetApplicationOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.Status)))

		return output, err
	}
	return nil, err
}

func waitApplicationDeleted(ctx context.Context, conn *qbusiness.QBusiness, id string, timeout time.Duration) (*qbusiness.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{qbusiness.ApplicationStatusActive, qbusiness.ApplicationStatusDeleting},
		Target:     []string{},
		Refresh:    statusAppAvailability(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qbusiness.GetApplicationOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.Status)))

		return output, err
	}
	return nil, err
}


func waitIndexCreated(ctx context.Context, conn *qbusiness.QBusiness, application_id, index_id string, timeout time.Duration) (*qbusiness.GetIndexOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{qbusiness.IndexStatusCreating, qbusiness.IndexStatusUpdating},
		Target:     []string{qbusiness.IndexStatusActive},
		Refresh:    statusIndexAvailability(ctx, conn, application_id, index_id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*qbusiness.GetIndexOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.Status)))

		return output, err
	}
	return nil, err
}
