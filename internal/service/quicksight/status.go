// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// status fetches the DataSource and its Status
func status(ctx context.Context, conn *quicksight.QuickSight, accountId, datasourceId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &quicksight.DescribeDataSourceInput{
			AwsAccountId: aws.String(accountId),
			DataSourceId: aws.String(datasourceId),
		}

		output, err := conn.DescribeDataSourceWithContext(ctx, input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.DataSource == nil {
			return nil, "", nil
		}

		return output.DataSource, aws.StringValue(output.DataSource.Status), nil
	}
}

// Fetch Template status
func statusTemplate(ctx context.Context, conn *quicksight.QuickSight, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindTemplateByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, *out.Version.Status, nil
	}
}

// Fetch Dashboard status
func statusDashboard(ctx context.Context, conn *quicksight.QuickSight, id string, version int64) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindDashboardByID(ctx, conn, id, version)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, *out.Version.Status, nil
	}
}

// Fetch Analysis status
func statusAnalysis(ctx context.Context, conn *quicksight.QuickSight, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindAnalysisByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, *out.Status, nil
	}
}

// Fetch Theme status
func statusTheme(ctx context.Context, conn *quicksight.QuickSight, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindThemeByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, *out.Version.Status, nil
	}
}
