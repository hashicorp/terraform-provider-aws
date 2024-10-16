// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicequotas

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findServiceQuotaDefaultByID(ctx context.Context, conn *servicequotas.Client, serviceCode, quotaCode string) (*types.ServiceQuota, error) {
	input := &servicequotas.GetAWSDefaultServiceQuotaInput{
		ServiceCode: aws.String(serviceCode),
		QuotaCode:   aws.String(quotaCode),
	}

	output, err := conn.GetAWSDefaultServiceQuota(ctx, input)

	if err != nil {
		return nil, err
	}
	if output == nil || output.Quota == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Quota, nil
}

func findServiceQuotaDefaultByName(ctx context.Context, conn *servicequotas.Client, serviceCode, quotaName string) (*types.ServiceQuota, error) {
	input := &servicequotas.ListAWSDefaultServiceQuotasInput{
		ServiceCode: aws.String(serviceCode),
	}

	paginator := servicequotas.NewListAWSDefaultServiceQuotasPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, q := range page.Quotas {
			if aws.ToString(q.QuotaName) == quotaName {
				return &q, nil
			}
		}
	}

	return nil, tfresource.NewEmptyResultError(input)
}

func findServiceQuotaByID(ctx context.Context, conn *servicequotas.Client, serviceCode, quotaCode string) (*types.ServiceQuota, error) {
	input := &servicequotas.GetServiceQuotaInput{
		ServiceCode: aws.String(serviceCode),
		QuotaCode:   aws.String(quotaCode),
	}

	output, err := conn.GetServiceQuota(ctx, input)

	var nsr *types.NoSuchResourceException
	if errors.As(err, &nsr) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || output.Quota == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if output.Quota.ErrorReason != nil {
		return nil, fmt.Errorf("%s: %s", output.Quota.ErrorReason.ErrorCode, aws.ToString(output.Quota.ErrorReason.ErrorMessage))
	}

	if output.Quota.Value == nil {
		return nil, &retry.NotFoundError{
			Message:     "empty value",
			LastRequest: input,
		}
	}

	return output.Quota, nil
}
