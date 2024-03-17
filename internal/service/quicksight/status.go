// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// status fetches the DataSource and its Status
func status(ctx context.Context, conn *quicksight.Client, accountId, datasourceId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &quicksight.DescribeDataSourceInput{
			AwsAccountId: aws.String(accountId),
			DataSourceId: aws.String(datasourceId),
		}

		output, err := conn.DescribeDataSource(ctx, input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.DataSource == nil {
			return nil, "", nil
		}
		return output.DataSource, flex.StringValueToFramework(ctx, output.DataSource.Status).String(), nil
	}
}

// Fetch Template status
func statusTemplate(ctx context.Context, conn *quicksight.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindTemplateByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, flex.StringValueToFramework(ctx, out.Version.Status).String(), nil
	}
}

// Fetch Dashboard status
func statusDashboard(ctx context.Context, conn *quicksight.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindDashboardByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, flex.StringValueToFramework(ctx, out.Version.Status).String(), nil
	}
}

// Fetch Analysis status
func statusAnalysis(ctx context.Context, conn *quicksight.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindAnalysisByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, flex.StringValueToFramework(ctx, out.Status).String(), nil
	}
}

// Fetch Theme status
func statusTheme(ctx context.Context, conn *quicksight.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindThemeByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, flex.StringValueToFramework(ctx, out.Version.Status).String(), nil
	}
}
