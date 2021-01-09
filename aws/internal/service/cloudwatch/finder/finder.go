package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

func CompositeAlarmByName(conn *cloudwatch.CloudWatch, name string) (*cloudwatch.CompositeAlarm, error) {
	input := cloudwatch.DescribeAlarmsInput{
		AlarmNames: aws.StringSlice([]string{name}),
		AlarmTypes: aws.StringSlice([]string{cloudwatch.AlarmTypeCompositeAlarm}),
	}

	output, err := conn.DescribeAlarms(&input)
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.CompositeAlarms) != 1 {
		return nil, nil
	}

	return output.CompositeAlarms[0], nil
}
