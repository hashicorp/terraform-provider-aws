package cloudwatch

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var validDashboardName = validation.All(
	validation.StringLenBetween(1, 255),
	validation.StringMatch(regexp.MustCompile(`^[\-_A-Za-z0-9]+$`), "must contain only alphanumeric characters, underscores and hyphens"),
)

var validEC2AutomateARN = validation.StringMatch(regexp.MustCompile(`^arn:[\w-]+:automate:[\w-]+:ec2:(reboot|recover|stop|terminate)$`), "must match EC2 automation ARN")
