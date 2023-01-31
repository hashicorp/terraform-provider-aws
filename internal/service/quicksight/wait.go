package quicksight

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	dataSourceCreateTimeout = 5 * time.Minute
	dataSourceUpdateTimeout = 5 * time.Minute
)

// waitCreated waits for a DataSource to return CREATION_SUCCESSFUL
func waitCreated(ctx context.Context, conn *quicksight.QuickSight, accountId, dataSourceId string) (*quicksight.DataSource, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{quicksight.ResourceStatusCreationInProgress},
		Target:  []string{quicksight.ResourceStatusCreationSuccessful},
		Refresh: status(ctx, conn, accountId, dataSourceId),
		Timeout: dataSourceCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*quicksight.DataSource); ok {
		if status, errorInfo := aws.StringValue(output.Status), output.ErrorInfo; status == quicksight.ResourceStatusCreationFailed && errorInfo != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(errorInfo.Type), aws.StringValue(errorInfo.Message)))
		}

		return output, err
	}

	return nil, err
}

// waitUpdated waits for a DataSource to return UPDATE_SUCCESSFUL
func waitUpdated(ctx context.Context, conn *quicksight.QuickSight, accountId, dataSourceId string) (*quicksight.DataSource, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{quicksight.ResourceStatusUpdateInProgress},
		Target:  []string{quicksight.ResourceStatusUpdateSuccessful},
		Refresh: status(ctx, conn, accountId, dataSourceId),
		Timeout: dataSourceUpdateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*quicksight.DataSource); ok {
		if status, errorInfo := aws.StringValue(output.Status), output.ErrorInfo; status == quicksight.ResourceStatusUpdateFailed && errorInfo != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(errorInfo.Type), aws.StringValue(errorInfo.Message)))
		}

		return output, err
	}

	return nil, err
}

// waitDataSetCreated waits for a DataSet to ensure at least one output column is created, proving success
func waitDataSetCreated(ctx context.Context, conn *quicksight.QuickSight, accountId, dataSetId string) (*quicksight.DataSet, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{quicksight.ResourceStatusCreationInProgress},
		Target:  []string{quicksight.ResourceStatusCreationSuccessful},
		Refresh: status(ctx, conn, accountId, dataSetId),
		Timeout: dataSourceCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*quicksight.DataSet); ok {
		if len(output.OutputColumns) == 0 {
			tfresource.SetLastError(err, fmt.Errorf("status check failed: failed to create quicksight data set"))
		}

		return output, err
	}

	return nil, err
}
