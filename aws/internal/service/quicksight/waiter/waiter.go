package waiter

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	DataSourceCreateTimeout = 5 * time.Minute
	DataSourceUpdateTimeout = 5 * time.Minute
)

// DataSourceCreated waits for a DataSource to return CREATION_SUCCESSFUL
func DataSourceCreated(ctx context.Context, conn *quicksight.QuickSight, accountId, dataSourceId string) (*quicksight.DataSource, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{quicksight.ResourceStatusCreationInProgress},
		Target:  []string{quicksight.ResourceStatusCreationSuccessful},
		Refresh: DataSourceStatus(ctx, conn, accountId, dataSourceId),
		Timeout: DataSourceCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*quicksight.DataSource); ok {
		if status, errorInfo := aws.StringValue(output.Status), output.ErrorInfo; status == quicksight.ResourceStatusCreationFailed && errorInfo != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(errorInfo.Type), aws.StringValue(errorInfo.Message)))
		}

		return output, err
	}

	return nil, err
}

// DataSourceUpdated waits for a DataSource to return UPDATE_SUCCESSFUL
func DataSourceUpdated(ctx context.Context, conn *quicksight.QuickSight, accountId, dataSourceId string) (*quicksight.DataSource, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{quicksight.ResourceStatusUpdateInProgress},
		Target:  []string{quicksight.ResourceStatusUpdateSuccessful},
		Refresh: DataSourceStatus(ctx, conn, accountId, dataSourceId),
		Timeout: DataSourceUpdateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*quicksight.DataSource); ok {
		if status, errorInfo := aws.StringValue(output.Status), output.ErrorInfo; status == quicksight.ResourceStatusUpdateFailed && errorInfo != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(errorInfo.Type), aws.StringValue(errorInfo.Message)))
		}

		return output, err
	}

	return nil, err
}
