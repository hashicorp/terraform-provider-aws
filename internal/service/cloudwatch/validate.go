package cloudwatch

import (
	"fmt"
	"regexp"
)

func validDashboardName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) > 255 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 255 characters: %q", k, value))
	}

	// http://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_PutDashboard.html
	pattern := `^[\-_A-Za-z0-9]+$`
	if !regexp.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q doesn't comply with restrictions (%q): %q",
			k, pattern, value))
	}

	return
}

func validEC2AutomateARN(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	// https://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_PutMetricAlarm.html
	pattern := `^arn:[\w-]+:automate:[\w-]+:ec2:(reboot|recover|stop|terminate)$`
	if !regexp.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q does not match EC2 automation ARN (%q): %q",
			k, pattern, value))
	}

	return
}
