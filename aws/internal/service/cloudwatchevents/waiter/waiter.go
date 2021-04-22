package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// ConnectionDeletedTimeout is the maximum amount of time to wait for a CloudwatchEvent Connection to delete
	ConnectionDeletedTimeout = 2 * time.Minute
)

// CloudWatchEventConnectionDeleted waits for a CloudwatchEvent Connection to return Deleted
func CloudWatchEventConnectionDeleted(conn *events.CloudWatchEvents, id string) (*events.DescribeConnectionOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{events.ConnectionStateDeleting},
		Target:  []string{},
		Refresh: CloudWatchEventConnectionStatus(conn, id),
		Timeout: ConnectionDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*events.DescribeConnectionOutput); ok {
		return v, err
	}

	return nil, err
}

// CloudWatchEventConnectionStatus fetches the Connection and its Status
func CloudWatchEventConnectionStatus(conn *events.CloudWatchEvents, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		params := events.DescribeConnectionInput{
			Name: aws.String(id),
		}

		output, err := conn.DescribeConnection(&params)
		if tfawserr.ErrMessageContains(err, events.ErrCodeResourceNotFoundException, "") {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ConnectionState), nil
	}
}
