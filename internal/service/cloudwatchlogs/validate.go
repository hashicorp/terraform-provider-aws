package cloudwatchlogs


import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/cognitoidentity"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tfnet "github.com/hashicorp/terraform-provider-aws/internal/net"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func validResourcePolicyDocument(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	// http://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_PutResourcePolicy.html
	if len(value) > 5120 || (len(value) == 0) {
		errors = append(errors, fmt.Errorf("CloudWatch log resource policy document must be between 1 and 5120 characters."))
	}
	if _, err := structure.NormalizeJsonString(v); err != nil {
		errors = append(errors, fmt.Errorf("%q contains an invalid JSON: %s", k, err))
	}
	return
}

func validLogGroupName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if len(value) > 512 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 512 characters: %q", k, value))
	}

	// http://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_CreateLogGroup.html
	pattern := `^[\.\-_/#A-Za-z0-9]+$`
	if !regexp.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q isn't a valid log group name (alphanumeric characters, underscores,"+
				" hyphens, slashes, hash signs and dots are allowed): %q",
			k, value))
	}

	return
}

func validLogGroupNamePrefix(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if len(value) > 483 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 483 characters: %q", k, value))
	}

	// http://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_CreateLogGroup.html
	pattern := `^[\.\-_/#A-Za-z0-9]+$`
	if !regexp.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q isn't a valid log group name (alphanumeric characters, underscores,"+
				" hyphens, slashes, hash signs and dots are allowed): %q",
			k, value))
	}

	return
}

func validLogMetricFilterName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if len(value) > 512 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 512 characters: %q", k, value))
	}

	// http://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_PutMetricFilter.html
	pattern := `^[^:*]+$`
	if !regexp.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q isn't a valid log metric name (must not contain colon nor asterisk): %q",
			k, value))
	}

	return
}

func validLogMetricFilterTransformationName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if len(value) > 255 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 255 characters: %q", k, value))
	}

	// http://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_MetricTransformation.html
	pattern := `^[^:*$]*$`
	if !regexp.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q isn't a valid log metric transformation name (must not contain"+
				" colon, asterisk nor dollar sign): %q",
			k, value))
	}

	return
}

