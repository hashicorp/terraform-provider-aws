package internetmonitor

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/internetmonitor"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func FindMonitor(ctx context.Context, conn *internetmonitor.InternetMonitor, name string) (*internetmonitor.GetMonitorOutput, error) {
	input := &internetmonitor.GetMonitorInput{
		MonitorName: aws.String(name),
	}

	output, err := conn.GetMonitorWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, internetmonitor.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
