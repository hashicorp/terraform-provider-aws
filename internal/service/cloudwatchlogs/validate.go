package cloudwatchlogs

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// http://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_PutResourcePolicy.html
var validResourcePolicyDocument = validation.All(
	validation.StringLenBetween(1, 5120),
	validation.StringIsJSON,
)

// http://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_CreateLogGroup.html
var validLogGroupName = validation.All(
	validation.StringLenBetween(1, 512),
	validation.StringMatch(regexp.MustCompile(`^[\.\-_/#A-Za-z0-9]+$`), "must contain only alphanumeric characters, underscores, hyphens, slashes, hash signs and periods"),
)

// http://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_CreateLogGroup.html
var validLogGroupNamePrefix = validation.All(
	validation.StringLenBetween(1, 483),
	validation.StringMatch(regexp.MustCompile(`^[\.\-_/#A-Za-z0-9]+$`), "must contain only alphanumeric characters, underscores, hyphens, slashes, hash signs and periods"),
)

// http://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_PutMetricFilter.html
var validLogMetricFilterName = validation.All(
	validation.StringLenBetween(1, 512),
	validation.StringMatch(regexp.MustCompile(`^[^:*]+$`), "must not contain a colon or an asterisk"),
)

// http://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_MetricTransformation.html
var validLogMetricFilterTransformationName = validation.All(
	validation.StringLenBetween(0, 255),
	validation.StringMatch(regexp.MustCompile(`^[^:*$]*$`), "must not contain a colon, an asterisk or a dollar sign"),
)
