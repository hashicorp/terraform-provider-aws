// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"fmt"
	"regexp"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	lazyvalidation "github.com/hashicorp/terraform-provider-aws/internal/sdkv2/validation/lazy"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	aliasNameCharPattern    = `[0-9A-Za-z_/-]+`
	aliasNamePattern        = `^` + aliasNamePrefix + aliasNameCharPattern + `$`
	grantNameCharPattern    = `[0-9A-Za-z_:/-]*`
	grantNamePattern        = `^` + grantNameCharPattern + `$`
	multiRegionKeyIDPattern = `mrk-[0-9a-f]{32}`
)

func aliasNameRegex() *regexp.Regexp {
	return regexache.MustCompile(aliasNamePattern)
}

var validGrantName = validation.All(
	validation.StringLenBetween(0, 256),
	lazyvalidation.StringMatch(grantNamePattern, "must only contain "+grantNameCharPattern),
)

var validNameForDataSource = lazyvalidation.StringMatch(
	aliasNamePattern,
	"must begin with 'alias/' and only contain "+aliasNameCharPattern,
)

var validNameForResource = validation.All(
	validation.StringDoesNotMatch(
		regexache.MustCompile(`^`+cmkAliasPrefix),
		"cannot begin with reserved AWS CMK prefix '"+cmkAliasPrefix+"'",
	),
	lazyvalidation.StringMatch(
		aliasNamePattern,
		"must begin with 'alias/' and only contain "+aliasNameCharPattern,
	),
)

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

var validateKeyID = lazyvalidation.StringMatch(`^`+verify.UUIDRegexPattern+`|`+multiRegionKeyIDPattern+`$`, "must be a KMS Key ID")

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
		errors = append(errors, fmt.Errorf("%q (%s) is an invalid ARN: %w", k, value, err))
		return
	}

	if !isKeyARN(value) {
		errors = append(errors, fmt.Errorf("%q (%s) is not a valid KMS Key ARN", k, value))
		return
	}

	return
}

var validateKeyAliasName = lazyvalidation.StringMatch(aliasNamePattern, "must be a KMS Key Alias")

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
		errors = append(errors, fmt.Errorf("%q (%s) is an invalid ARN: %w", k, value, err))
		return
	}

	if !isAliasARN(value) {
		errors = append(errors, fmt.Errorf("%q (%s) is not a valid KMS Key Alias ARN", k, value))
		return
	}

	return
}
