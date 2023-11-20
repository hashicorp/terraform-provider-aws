// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/apprunner"
	"github.com/aws/aws-sdk-go-v2/service/apprunner/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

const (
	ServiceCreateTimeout = 20 * time.Minute
	ServiceDeleteTimeout = 20 * time.Minute
	ServiceUpdateTimeout = 20 * time.Minute
)

func WaitServiceCreated(ctx context.Context, conn *apprunner.Client, serviceArn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ServiceStatusOperationInProgress),
		Target:  enum.Slice(types.ServiceStatusRunning),
		Refresh: StatusService(ctx, conn, serviceArn),
		Timeout: ServiceCreateTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitServiceUpdated(ctx context.Context, conn *apprunner.Client, serviceArn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ServiceStatusOperationInProgress),
		Target:  enum.Slice(types.ServiceStatusRunning),
		Refresh: StatusService(ctx, conn, serviceArn),
		Timeout: ServiceUpdateTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func WaitServiceDeleted(ctx context.Context, conn *apprunner.Client, serviceArn string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ServiceStatusRunning, types.ServiceStatusOperationInProgress),
		Target:  enum.Slice(types.ServiceStatusDeleted),
		Refresh: StatusService(ctx, conn, serviceArn),
		Timeout: ServiceDeleteTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
