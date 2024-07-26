// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	aliasNamePattern        = aliasNamePrefix + `[0-9A-Za-z_/-]+`
	multiRegionKeyIDPattern = `mrk-[0-9a-f]{32}`
)

var (
	aliasNameRegex     = regexache.MustCompile(`^` + aliasNamePattern + `$`)
	keyIDRegex         = regexache.MustCompile(`^` + verify.UUIDRegexPattern + `|` + multiRegionKeyIDPattern + `$`)
	keyIDResourceRegex = regexache.MustCompile(`^key/(` + verify.UUIDRegexPattern + `|` + multiRegionKeyIDPattern + `)$`)
)

func validGrantName(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)

	if len(value) > 256 {
		es = append(es, fmt.Errorf("%s can not be greater than 256 characters", k))
	}

	if !regexache.MustCompile(`^[0-9A-Za-z_:/-]+$`).MatchString(value) {
		es = append(es, fmt.Errorf("%s must only contain [0-9A-Za-z_:/-]", k))
	}

	return
}

func validNameForDataSource(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)

	if !aliasNameRegex.MatchString(value) {
		es = append(es, fmt.Errorf(
			"%q must begin with 'alias/' and be comprised of only [0-9A-Za-z_/-]", k))
	}
	return
}

func validNameForResource(v interface{}, k string) (ws []string, es []error) {
	value := v.(string)

	if regexache.MustCompile(`^(` + cmkAliasPrefix + `)`).MatchString(value) {
		es = append(es, fmt.Errorf("%q cannot begin with reserved AWS CMK prefix 'alias/aws/'", k))
	}

	if !aliasNameRegex.MatchString(value) {
		es = append(es, fmt.Errorf(
			"%q must begin with 'alias/' and be comprised of only [0-9A-Za-z_/-]", k))
	}
	return
}

var validateKey = validation.Any(
	validateKeyID,
	validateKeyARN,
)

var validateKeyOrAlias = validation.Any(
	validateKeyID,
	validateKeyARN,
	validateKeyAliasName,
	validateKeyAliasARN,
)

var validateKeyID = validation.StringMatch(keyIDRegex, "must be a KMS Key ID")

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
