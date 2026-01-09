// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	functionNameRegex = regexache.MustCompile("(arn:(aws[a-zA-Z-]*)?:lambda:)?([a-z]{2}((-gov)|(-iso([a-z]?)))?-[a-z]+-\\d{1}:)?(\\d{12}:)?(function:)?([a-zA-Z0-9-_]+)")
)

var functionNameValidator validator.String = stringvalidator.RegexMatches(functionNameRegex, "must be a valid function name")

func validFunctionName() schema.SchemaValidateFunc {
	// http://docs.aws.amazon.com/lambda/latest/dg/API_AddPermission.html
	pattern := `^(arn:[\w-]+:lambda:)?([a-z]{2}-(?:[a-z]+-){1,2}\d{1}:)?(\d{12}:)?(function:)?([0-9A-Za-z_-]+)(:(\$LATEST|[0-9A-Za-z_-]+))?$`

	fullARNRegex := regexache.MustCompile(`^arn:[\w-]+:lambda:[a-z]{2}-(?:[a-z]+-){1,2}\d{1}:\d{12}:function:([0-9A-Za-z_-]+)(:(\$LATEST|[0-9A-Za-z_-]+))?$`)
	partialARNRegex := regexache.MustCompile(`^\d{12}:function:([0-9A-Za-z_-]+)$`)

	return validation.All(
		validation.StringMatch(regexache.MustCompile(pattern), "must be valid function name or function ARN"),
		validation.StringLenBetween(1, 140),
		func(i any, k string) (warnings []string, errors []error) {
			v, ok := i.(string)
			if !ok {
				errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
				return warnings, errors
			}

			var functionName string
			if matched := fullARNRegex.FindStringSubmatch(v); matched != nil {
				functionName = matched[1]
			} else if matched := partialARNRegex.FindStringSubmatch(v); matched != nil {
				functionName = matched[1]
			} else {
				functionName = v
			}
			if len(functionName) > 64 {
				errors = append(errors, fmt.Errorf("expected length of the Lambda function name to be less than or equal to 64, got %s (%d characters)", functionName, len(functionName)))
			}

			return warnings, errors
		},
	)
}

func validPermissionAction() schema.SchemaValidateFunc {
	pattern := `^(lambda:[*]|lambda:[A-Za-z]+|[*])$`
	return validation.StringMatch(regexache.MustCompile(pattern), "must be a valid action (usually starts with lambda:)")
}

func validPermissionEventSourceToken() schema.SchemaValidateFunc {
	// https://docs.aws.amazon.com/lambda/latest/dg/API_AddPermission.html
	return validation.All(
		validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]+$`), "must contain alphanumeric, periods, underscores or dashes only"),
		validation.StringLenBetween(1, 256),
	)
}

func validQualifier() schema.SchemaValidateFunc {
	// http://docs.aws.amazon.com/lambda/latest/dg/API_AddPermission.html
	return validation.All(
		validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_$-]+$`), "must contain alphanumeric, dollar signs, underscores or dashes only"),
		validation.StringLenBetween(1, 128),
	)
}

func validPolicyStatementID() schema.SchemaValidateFunc {
	// http://docs.aws.amazon.com/lambda/latest/dg/API_AddPermission.html
	return validation.All(
		validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_-]+$`), "must contain alphanumeric, underscores or dashes only"),
		validation.StringLenBetween(1, 100),
	)
}

func validLogGroupName() schema.SchemaValidateFunc {
	// http://docs.aws.amazon.com/AmazonCloudWatchLogs/latest/APIReference/API_CreateLogGroup.html
	return validation.All(
		validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_./#-]+$`), "must contain alphanumeric characters, underscores,"+
			" hyphens, slashes, hash signs and dots only"),
		validation.StringLenBetween(1, 512),
	)
}
