package waiter

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// DataSourceStatus fetches the DataSource and its Status
func DataSourceStatus(ctx context.Context, conn *quicksight.QuickSight, accountId, datasourceId string) resource.StateRefreshFunc {
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
