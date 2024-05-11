// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func waitServiceSettingUpdated(ctx context.Context, conn *ssm.SSM, id string, timeout time.Duration) (*ssm.ServiceSetting, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"PendingUpdate", ""},
		Target:  []string{"Customized", "Default"},
		Refresh: statusServiceSetting(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ssm.ServiceSetting); ok {
		return output, err
	}

	return nil, err
}

func waitServiceSettingReset(ctx context.Context, conn *ssm.SSM, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"Customized", "PendingUpdate", ""},
		Target:  []string{"Default"},
		Refresh: statusServiceSetting(ctx, conn, id),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
