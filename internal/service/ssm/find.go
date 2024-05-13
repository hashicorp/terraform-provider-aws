// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindServiceSettingByID(ctx context.Context, conn *ssm.SSM, id string) (*ssm.ServiceSetting, error) {
	input := &ssm.GetServiceSettingInput{
		SettingId: aws.String(id),
	}

	output, err := conn.GetServiceSettingWithContext(ctx, input)

	if tfawserr.ErrCodeContains(err, ssm.ErrCodeServiceSettingNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ServiceSetting == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ServiceSetting, nil
}
