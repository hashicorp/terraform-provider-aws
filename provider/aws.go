package provider

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func validateArn(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if value == "" {
		return ws, errors
	}

	parsedARN, err := arn.Parse(value)

	if err != nil {
		errors = append(errors, fmt.Errorf("%q (%s) is an invalid ARN: %s", k, value, err))
		return ws, errors
	}

	if parsedARN.Partition == "" {
		errors = append(errors, fmt.Errorf("%q (%s) is an invalid ARN: missing partition value", k, value))
	} else if !PartitionRegexp.MatchString(parsedARN.Partition) {
		errors = append(errors, fmt.Errorf("%q (%s) is an invalid ARN: invalid partition value (expecting to match regular expression: %s)", k, value, PartitionRegexpPattern))
	}

	if parsedARN.Region != "" && !RegionRegexp.MatchString(parsedARN.Region) {
		errors = append(errors, fmt.Errorf("%q (%s) is an invalid ARN: invalid region value (expecting to match regular expression: %s)", k, value, RegionRegexpPattern))
	}

	if parsedARN.AccountID != "" && !AccountIDRegexp.MatchString(parsedARN.AccountID) {
		errors = append(errors, fmt.Errorf("%q (%s) is an invalid ARN: invalid account ID value (expecting to match regular expression: %s)", k, value, AccountIDRegexpPattern))
	}

	if parsedARN.Resource == "" {
		errors = append(errors, fmt.Errorf("%q (%s) is an invalid ARN: missing resource value", k, value))
	}

	return ws, errors
}

func EC2RegionalPrivateDnsSuffix(region string) string {
	if region == endpoints.UsEast1RegionID {
		return "ec2.internal"
	}

	return fmt.Sprintf("%s.compute.internal", region)
}

func EC2DashIP(ip string) string {
	return strings.Replace(ip, ".", "-", -1)
}

const (
	PartitionRegexpInternalPattern     = `aws(-[a-z]+)*`
	AccountIDRegexpInternalPattern     = `(aws|\d{12})`
	RegionRegexpInternalPattern        = `[a-z]{2}(-[a-z]+)+-\d`
	VersionStringRegexpInternalPattern = `[[:digit:]]+(\.[[:digit:]]+){2}`
)

const (
	AccountIDRegexpPattern     = "^" + AccountIDRegexpInternalPattern + "$"
	PartitionRegexpPattern     = "^" + PartitionRegexpInternalPattern + "$"
	RegionRegexpPattern        = "^" + RegionRegexpInternalPattern + "$"
	VersionStringRegexpPattern = "^" + VersionStringRegexpInternalPattern + "$"
)

var AccountIDRegexp = regexp.MustCompile(AccountIDRegexpPattern)
var PartitionRegexp = regexp.MustCompile(PartitionRegexpPattern)
var RegionRegexp = regexp.MustCompile(RegionRegexpPattern)

var VersionStringRegexp = regexp.MustCompile(VersionStringRegexpPattern)

func HasEC2Classic(platforms []string) bool {
	for _, p := range platforms {
		if p == "EC2" {
			return true
		}
	}
	return false
}

// ReverseDNS switches a DNS hostname to reverse DNS and vice-versa.
func ReverseDNS(hostname string) string {
	parts := strings.Split(hostname, ".")

	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	return strings.Join(parts, ".")
}

func SupportedEC2Platforms(conn *ec2.EC2) ([]string, error) {
	attrName := "supported-platforms"

	input := ec2.DescribeAccountAttributesInput{
		AttributeNames: []*string{aws.String(attrName)},
	}
	attributes, err := conn.DescribeAccountAttributes(&input)
	if err != nil {
		return nil, err
	}

	var platforms []string
	for _, attr := range attributes.AccountAttributes {
		if *attr.AttributeName == attrName {
			for _, v := range attr.AttributeValues {
				platforms = append(platforms, *v.AttributeValue)
			}
			break
		}
	}

	if len(platforms) == 0 {
		return nil, fmt.Errorf("No EC2 platforms detected")
	}

	return platforms, nil
}
