// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package qbusiness

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/qbusiness"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func waitApplicationCreated(ctx context.Context, conn *qbusiness.QBusiness, id string, timeout time.Duration) (*redshift.Cluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{applicationStatusUpdating, applicationStatusCreating},
		Target:     []string{applicationStatusActive},
		Refresh:    statusAppAvailability(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshift.Cluster); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.ClusterStatus)))

		return output, err
	}
	return nil, err
}
