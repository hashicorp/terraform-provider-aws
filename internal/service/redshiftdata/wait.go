package redshiftdata

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftdataapiservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func waitStatementFinished(conn *redshiftdataapiservice.RedshiftDataAPIService, id string, timeout time.Duration) (*redshiftdataapiservice.DescribeStatementOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			redshiftdataapiservice.StatusStringPicked,
			redshiftdataapiservice.StatusStringStarted,
			redshiftdataapiservice.StatusStringSubmitted,
		},
		Target:     []string{redshiftdataapiservice.StatusStringFinished},
		Refresh:    statusStatement(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*redshiftdataapiservice.DescribeStatementOutput); ok {
		if status := aws.StringValue(output.Status); status == redshiftdataapiservice.StatusStringFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.Error)))
		}

		return output, err
	}

	return nil, err
}
