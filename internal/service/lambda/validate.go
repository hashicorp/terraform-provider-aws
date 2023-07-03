// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func validFunctionName() schema.SchemaValidateFunc {
	// http://docs.aws.amazon.com/lambda/latest/dg/API_AddPermission.html
	pattern := `^(arn:[\w-]+:lambda:)?([a-z]{2}-(?:[a-z]+-){1,2}\d{1}:)?(\d{12}:)?(function:)?([a-zA-Z0-9-_]+)(:(\$LATEST|[a-zA-Z0-9-_]+))?$`

	return validation.All(
		validation.StringMatch(regexp.MustCompile(pattern), "must be valid function name or function ARN"),
		validation.StringLenBetween(1, 140),
	)
}

func validPermissionAction() schema.SchemaValidateFunc {
	pattern := `^(lambda:[*]|lambda:[a-zA-Z]+|[*])$`
	return validation.StringMatch(regexp.MustCompile(pattern), "must be a valid action (usually starts with lambda:)")
}

func validPermissionEventSourceToken() schema.SchemaValidateFunc {
	// https://docs.aws.amazon.com/lambda/latest/dg/API_AddPermission.html
	return validation.All(
		validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9._\-]+$`), "must contain alphanumeric, periods, underscores or dashes only"),
		validation.StringLenBetween(1, 256),
	)
}

func validQualifier() schema.SchemaValidateFunc {
	// http://docs.aws.amazon.com/lambda/latest/dg/API_AddPermission.html
	return validation.All(
		validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9$_-]+$`), "must contain alphanumeric, dollar signs, underscores or dashes only"),
		validation.StringLenBetween(1, 128),
	)
}

func validPolicyStatementID() schema.SchemaValidateFunc {
	// http://docs.aws.amazon.com/lambda/latest/dg/API_AddPermission.html
	return validation.All(
		validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9_-]+$`), "must contain alphanumeric, underscores or dashes only"),
		validation.StringLenBetween(1, 100),
	)
}
