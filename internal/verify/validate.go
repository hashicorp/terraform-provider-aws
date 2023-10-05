// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verify

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws/arn"
	basevalidation "github.com/hashicorp/aws-sdk-go-base/v2/validation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/types/timestamp"
)

var accountIDRegexp = regexache.MustCompile(`^(aws|aws-managed|third-party|\d{12}|cw.{10})$`)
var partitionRegexp = regexache.MustCompile(`^aws(-[a-z]+)*$`)
var regionRegexp = regexache.MustCompile(`^[a-z]{2}(-[a-z]+)+-\d$`)

// validates all listed in https://gist.github.com/shortjared/4c1e3fe52bdfa47522cfe5b41e5d6f22
var servicePrincipalRegexp = regexache.MustCompile(`^([0-9a-z-]+\.){1,4}(amazonaws|amazon)\.com$`)

func Valid4ByteASN(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	asn, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q (%q) must be a 64-bit integer", k, v))
		return
	}

	if asn < 0 || asn > 4294967295 {
		errors = append(errors, fmt.Errorf("%q (%q) must be in the range 0 to 4294967295", k, v))
	}
	return
}

func ValidAmazonSideASN(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	// http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateVpnGateway.html
	asn, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q (%q) must be a 64-bit integer", k, v))
		return
	}

	// https://github.com/hashicorp/terraform-provider-aws/issues/5263
	isLegacyAsn := func(a int64) bool {
		return a == 7224 || a == 9059 || a == 10124 || a == 17493
	}

	if !isLegacyAsn(asn) && ((asn < 64512) || (asn > 65534 && asn < 4200000000) || (asn > 4294967294)) {
		errors = append(errors, fmt.Errorf("%q (%q) must be 7224, 9059, 10124 or 17493 or in the range 64512 to 65534 or 4200000000 to 4294967294", k, v))
	}
	return
}

// ValidARN validates that a string value matches a generic ARN format
var ValidARN = ValidARNCheck()

type ARNCheckFunc func(any, string, arn.ARN) ([]string, []error)

// ValidARNCheck validates that a string value matches an ARN format with additional validation on the parsed ARN value
// It must:
// * Be parseable as an ARN
// * Have a valid partition
// * Have a valid region
// * Have either an empty or valid account ID
// * Have a non-empty resource part
// * Pass the supplied checks
func ValidARNCheck(f ...ARNCheckFunc) schema.SchemaValidateFunc {
	return func(v any, k string) (ws []string, errors []error) {
		value, ok := v.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
			return ws, errors
		}

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
		} else if !partitionRegexp.MatchString(parsedARN.Partition) {
			errors = append(errors, fmt.Errorf("%q (%s) is an invalid ARN: invalid partition value (expecting to match regular expression: %s)", k, value, partitionRegexp))
		}

		if parsedARN.Region != "" && !regionRegexp.MatchString(parsedARN.Region) {
			errors = append(errors, fmt.Errorf("%q (%s) is an invalid ARN: invalid region value (expecting to match regular expression: %s)", k, value, regionRegexp))
		}

		if parsedARN.AccountID != "" && !accountIDRegexp.MatchString(parsedARN.AccountID) {
			errors = append(errors, fmt.Errorf("%q (%s) is an invalid ARN: invalid account ID value (expecting to match regular expression: %s)", k, value, accountIDRegexp))
		}

		if parsedARN.Resource == "" {
			errors = append(errors, fmt.Errorf("%q (%s) is an invalid ARN: missing resource value", k, value))
		}

		for _, f := range f {
			w, e := f(v, k, parsedARN)
			ws = append(ws, w...)
			errors = append(errors, e...)
		}

		return ws, errors
	}
}

func ValidAccountID(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	// http://docs.aws.amazon.com/lambda/latest/dg/API_AddPermission.html
	pattern := `^\d{12}$`
	if !regexache.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q doesn't look like AWS Account ID (exactly 12 digits): %q",
			k, value))
	}

	return
}

// ValidCIDRNetworkAddress ensures that the string value is a valid CIDR that
// represents a network address - it adds an error otherwise
func ValidCIDRNetworkAddress(v interface{}, k string) (ws []string, errors []error) {
	if err := types.ValidateCIDRBlock(v.(string)); err != nil {
		errors = append(errors, err)
		return
	}

	return
}

func ValidIAMPolicyJSON(v interface{}, k string) (ws []string, errors []error) {
	// IAM Policy documents need to be valid JSON, and pass legacy parsing
	value := v.(string)
	if len(value) < 1 {
		errors = append(errors, fmt.Errorf("%q is an empty string, which is not a valid JSON value", k))
	} else if first := value[:1]; first != "{" {
		switch value[:1] {
		case " ", "\t", "\r", "\n":
			errors = append(errors, fmt.Errorf("%q contains an invalid JSON policy: leading space characters are not allowed", k))
		case `"`:
			// There are some common mistakes that lead to strings appearing
			// here instead of objects, so we'll try some heuristics to
			// check for those so we might give more actionable feedback in
			// these situations.
			var hint string
			var content string
			var innerContent any
			if err := json.Unmarshal([]byte(value), &content); err == nil {
				if strings.HasSuffix(content, ".json") {
					hint = " (have you passed a JSON-encoded filename instead of the content of that file?)"
				} else if err := json.Unmarshal([]byte(content), &innerContent); err == nil {
					hint = " (have you double-encoded your JSON data?)"
				}
			}
			errors = append(errors, fmt.Errorf("%q contains an invalid JSON policy: contains a JSON-encoded string, not a JSON-encoded object%s", k, hint))
		case `[`:
			errors = append(errors, fmt.Errorf("%q contains an invalid JSON policy: contains a JSON array, not a JSON object", k))
		default:
			// Generic error for if we didn't find something more specific to say.
			errors = append(errors, fmt.Errorf("%q contains an invalid JSON policy: not a JSON object", k))
		}
	} else if _, err := structure.NormalizeJsonString(v); err != nil {
		errStr := err.Error()
		if err, ok := errs.As[*json.SyntaxError](err); ok {
			errStr = fmt.Sprintf("%s, at byte offset %d", errStr, err.Offset)
		}
		errors = append(errors, fmt.Errorf("%q contains an invalid JSON policy: %s", k, errStr))
	} else if err := basevalidation.JSONNoDuplicateKeys(value); err != nil {
		errors = append(errors, fmt.Errorf("%q contains duplicate JSON keys: %s", k, err))
	}

	return //nolint:nakedret // Just a long function.
}

// ValidateIPv4CIDRBlock validates that the specified CIDR block is valid:
// - The CIDR block parses to an IP address and network
// - The IP address is an IPv4 address
// - The CIDR block is the CIDR block for the network
func ValidateIPv4CIDRBlock(cidr string) error {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("%q is not a valid CIDR block: %w", cidr, err)
	}

	ipv4 := ip.To4()
	if ipv4 == nil {
		return fmt.Errorf("%q is not a valid IPv4 CIDR block", cidr)
	}

	if !types.CIDRBlocksEqual(cidr, ipnet.String()) {
		return fmt.Errorf("%q is not a valid IPv4 CIDR block; did you mean %q?", cidr, ipnet)
	}

	return nil
}

// ValidateIPv6CIDRBlock validates that the specified CIDR block is valid:
// - The CIDR block parses to an IP address and network
// - The IP address is an IPv6 address
// - The CIDR block is the CIDR block for the network
func ValidateIPv6CIDRBlock(cidr string) error {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("%q is not a valid CIDR block: %w", cidr, err)
	}

	ipv4 := ip.To4()
	if ipv4 != nil {
		return fmt.Errorf("%q is not a valid IPv6 CIDR block", cidr)
	}

	if !types.CIDRBlocksEqual(cidr, ipnet.String()) {
		return fmt.Errorf("%q is not a valid IPv6 CIDR block; did you mean %q?", cidr, ipnet)
	}

	return nil
}

// ValidIPv4CIDRNetworkAddress ensures that the string value is a valid IPv4 CIDR that
// represents a network address - it adds an error otherwise
func ValidIPv4CIDRNetworkAddress(v interface{}, k string) (ws []string, errors []error) {
	if err := ValidateIPv4CIDRBlock(v.(string)); err != nil {
		errors = append(errors, err)
		return
	}

	return
}

// ValidIPv6CIDRNetworkAddress ensures that the string value is a valid IPv6 CIDR that
// represents a network address - it adds an error otherwise
func ValidIPv6CIDRNetworkAddress(v interface{}, k string) (ws []string, errors []error) {
	if err := ValidateIPv6CIDRBlock(v.(string)); err != nil {
		errors = append(errors, err)
		return
	}

	return
}

// IsIPv4CIDRBlockOrIPv6CIDRBlock returns a SchemaValidateFunc that test if the provided value:
// - Is a valid IPv4 CIDR block and passes the specified validation, or
// - Is a valid IPv6 CIDR block and passes the specified validation
func IsIPv4CIDRBlockOrIPv6CIDRBlock(ipv4Validator, ipv6Validator schema.SchemaValidateFunc) schema.SchemaValidateFunc {
	return validation.Any(
		validation.All(ValidIPv4CIDRNetworkAddress, ipv4Validator),
		validation.All(ValidIPv6CIDRNetworkAddress, ipv6Validator),
	)
}

// KMS Key IDs (a subset of KMS Key Identifiers) can be be key ID, key ARN, alias name, or alias ARN.
// There's no guarantee about the format of a Key ID other than a string between 1 and 2048 characters
// (per KMS API documentation and internal AWS conversations).
// ref: https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#key-id
// ref: https://docs.aws.amazon.com/kms/latest/APIReference/API_Encrypt.html#KMS-Encrypt-request-KeyId
func ValidKMSKeyID(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) < 1 {
		errors = append(errors, fmt.Errorf("%q cannot be shorter than 1 character", k))
	} else if len(value) > 2048 {
		errors = append(errors, fmt.Errorf("%q cannot be longer than 2048 characters", k))
	}
	return
}

func ValidLaunchTemplateID(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) < 1 {
		errors = append(errors, fmt.Errorf("%q cannot be shorter than 1 character", k))
	} else if len(value) > 255 {
		errors = append(errors, fmt.Errorf("%q cannot be longer than 255 characters", k))
	} else if !regexache.MustCompile(`^lt\-[0-9a-z]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q must begin with 'lt-' and be comprised of only alphanumeric characters: %v", k, value))
	}
	return
}

func ValidLaunchTemplateName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) < 3 {
		errors = append(errors, fmt.Errorf("%q cannot be less than 3 characters", k))
	} else if strings.HasSuffix(k, "prefix") && len(value) > 99 {
		errors = append(errors, fmt.Errorf("%q cannot be longer than 99 characters, name is limited to 125", k))
	} else if !strings.HasSuffix(k, "prefix") && len(value) > 125 {
		errors = append(errors, fmt.Errorf("%q cannot be longer than 125 characters", k))
	} else if !regexache.MustCompile(`^[0-9A-Za-z()./_\-]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf("%q can only alphanumeric characters and ()./_- symbols", k))
	}
	return
}

// validateMulticastIPAddress validates that the specified string is a multicast IP address.
func validateMulticastIPAddress(s string) error {
	ip := net.ParseIP(s)
	if ip == nil {
		return fmt.Errorf("%q is not a valid IP address", s)
	}

	if !ip.IsMulticast() {
		return fmt.Errorf("%q is not a valid multicast address", s)
	}

	return nil
}

func ValidMulticastIPAddress(v interface{}, k string) (ws []string, errors []error) {
	if err := validateMulticastIPAddress(v.(string)); err != nil {
		errors = append(errors, err)
		return
	}

	return
}

func ValidOnceADayWindowFormat(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	t := timestamp.New(value)
	if err := t.ValidateOnceADayWindowFormat(); err != nil {
		errors = append(errors, err)
		return
	}

	return
}

func ValidOnceAWeekWindowFormat(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	t := timestamp.New(value)
	if err := t.ValidateOnceAWeekWindowFormat(); err != nil {
		errors = append(errors, err)
		return
	}

	return
}

func ValidRegionName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if value == "" {
		return ws, errors
	}
	if !regionRegexp.MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q region name is malformed(%q): %q",
			k, regionRegexp, value))
	}

	return
}

func ValidStringIsJSONOrYAML(v interface{}, k string) (ws []string, errors []error) {
	if looksLikeJSONString(v) {
		if _, err := structure.NormalizeJsonString(v); err != nil {
			errors = append(errors, fmt.Errorf("%q contains an invalid JSON: %s", k, err))
		}
	} else {
		if _, err := checkYAMLString(v); err != nil {
			errors = append(errors, fmt.Errorf("%q contains an invalid YAML: %s", k, err))
		}
	}
	return
}

// ValidTypeStringNullableFloat provides custom error messaging for TypeString floats
// Some arguments require a floating point value or an unspecified, empty field.
func ValidTypeStringNullableFloat(v interface{}, k string) (ws []string, es []error) {
	value, ok := v.(string)
	if !ok {
		es = append(es, fmt.Errorf("expected type of %s to be string", k))
		return
	}

	if value == "" {
		return
	}

	if _, err := strconv.ParseFloat(value, 64); err != nil {
		es = append(es, fmt.Errorf("%s: cannot parse '%s' as float: %s", k, value, err))
	}

	return
}

// ValidUTCTimestamp validates a string in UTC Format required by APIs including:
// https://docs.aws.amazon.com/iot/latest/apireference/API_CloudwatchMetricAction.html
// https://docs.aws.amazon.com/AmazonRDS/latest/APIReference/API_RestoreDBInstanceToPointInTime.html
func ValidUTCTimestamp(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	t := timestamp.New(value)
	if err := t.ValidateUTCFormat(); err != nil {
		errors = append(errors, err)
		return
	}

	return
}

var ValidStringDateOrPositiveInt = validation.Any(
	validation.IsRFC3339Time,
	validation.StringMatch(regexache.MustCompile(`^\d+$`), "must be a positive integer value"),
)

func ValidDuration(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	duration, err := time.ParseDuration(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q cannot be parsed as a duration: %s", k, err))
	}
	if duration < 0 {
		errors = append(errors, fmt.Errorf("%q must be greater than zero", k))
	}
	return
}

// FloatGreaterThan returns a SchemaValidateFunc which tests if the provided value
// is of type float and is greater than threshold.
func FloatGreaterThan(threshold float64) schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(float64)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be float", k))
			return
		}

		if v <= threshold {
			es = append(es, fmt.Errorf("expected %s to be greater than (%f), got %f", k, threshold, v))
			return
		}

		return
	}
}

func ValidServicePrincipal(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if value == "" {
		return ws, errors
	}

	if !IsServicePrincipal(value) {
		errors = append(errors, fmt.Errorf("%q (%s) is an invalid Service Principal: invalid prefix value (expecting to match regular expression: %s)", k, value, servicePrincipalRegexp))
	}

	return ws, errors
}

func IsServicePrincipal(value string) (valid bool) {
	return servicePrincipalRegexp.MatchString(value)
}
