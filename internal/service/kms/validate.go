// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	aliasNameRegexPattern   = `alias/[a-zA-Z0-9/_-]+`
	multiRegionKeyIdPattern = `mrk-[a-f0-9]{32}`
)

var (
	aliasNameRegex     = regexp.MustCompile(`^` + aliasNameRegexPattern + `$`)
	keyIdRegex         = regexp.MustCompile(`^` + verify.UUIDRegexPattern + `|` + multiRegionKeyIdPattern + `$`)
	keyIdResourceRegex = regexp.MustCompile(`^key/(` + verify.UUIDRegexPattern + `|` + multiRegionKeyIdPattern + `)$`)
)

func validGrantName(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)

	if len(value) > 256 {
		es = append(es, fmt.Errorf("%s can not be greater than 256 characters", k))
	}

	if !regexp.MustCompile(`^[a-zA-Z0-9:/_-]+$`).MatchString(value) {
		es = append(es, fmt.Errorf("%s must only contain [a-zA-Z0-9:/_-]", k))
	}

	return
}

func validNameForDataSource(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)

	if !aliasNameRegex.MatchString(value) {
		es = append(es, fmt.Errorf(
			"%q must begin with 'alias/' and be comprised of only [a-zA-Z0-9/_-]", k))
	}
	return
}

func validNameForResource(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)

	if regexp.MustCompile(`^(alias/aws/)`).MatchString(value) {
		es = append(es, fmt.Errorf("%q cannot begin with reserved AWS CMK prefix 'alias/aws/'", k))
	}

	if !aliasNameRegex.MatchString(value) {
		es = append(es, fmt.Errorf(
			"%q must begin with 'alias/' and be comprised of only [a-zA-Z0-9/_-]", k))
	}
	return
}

var ValidateKey = validation.Any(
	validateKeyId,
	validateKeyARN,
)

var ValidateKeyOrAlias = validation.Any(
	validateKeyId,
	validateKeyARN,
	validateKeyAliasName,
	validateKeyAliasARN,
)

var validateKeyId = validation.StringMatch(keyIdRegex, "must be a KMS Key ID")

func validateKeyARN(v any, k string) (ws []string, errors []error) {
	value, ok := v.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return
	}

	if value == "" {
		return
	}

	if _, err := arn.Parse(value); err != nil {
		errors = append(errors, fmt.Errorf("%q (%s) is an invalid ARN: %s", k, value, err))
		return
	}

	if !isKeyARN(value) {
		errors = append(errors, fmt.Errorf("%q (%s) is not a valid KMS Key ARN", k, value))
		return
	}

	return
}

var validateKeyAliasName = validation.StringMatch(aliasNameRegex, "must be a KMS Key Alias")

func validateKeyAliasARN(v any, k string) (ws []string, errors []error) {
	value, ok := v.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return
	}

	if value == "" {
		return
	}

	if _, err := arn.Parse(value); err != nil {
		errors = append(errors, fmt.Errorf("%q (%s) is an invalid ARN: %s", k, value, err))
		return
	}

	if !isAliasARN(value) {
		errors = append(errors, fmt.Errorf("%q (%s) is not a valid KMS Key Alias ARN", k, value))
		return
	}

	return
}
