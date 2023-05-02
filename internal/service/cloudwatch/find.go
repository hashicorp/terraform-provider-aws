package cloudwatch

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

func FindMetricAlarmByName(ctx context.Context, conn *cloudwatch.CloudWatch, name string) (*cloudwatch.MetricAlarm, error) {
	input := cloudwatch.DescribeAlarmsInput{
		AlarmNames: []*string{aws.String(name)},
		AlarmTypes: aws.StringSlice([]string{cloudwatch.AlarmTypeMetricAlarm}),
	}

	output, err := conn.DescribeAlarmsWithContext(ctx, &input)
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.MetricAlarms) != 1 {
		return nil, nil
	}

	return output.MetricAlarms[0], nil
}
