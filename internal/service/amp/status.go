// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusAlertManagerDefinition(ctx context.Context, conn *prometheusservice.PrometheusService, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindAlertManagerDefinitionByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.StatusCode), nil
	}
}

func statusRuleGroupNamespace(ctx context.Context, conn *prometheusservice.PrometheusService, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindRuleGroupNamespaceByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.StatusCode), nil
	}
}

func statusWorkspace(ctx context.Context, conn *prometheusservice.PrometheusService, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindWorkspaceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.StatusCode), nil
	}
}

func statusLoggingConfiguration(ctx context.Context, conn *prometheusservice.PrometheusService, workspaceID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindLoggingConfigurationByWorkspaceID(ctx, conn, workspaceID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.StatusCode), nil
	}
}
