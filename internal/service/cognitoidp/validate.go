// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"fmt"
	"unicode/utf8"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

func validIdentityProviderName(v any, k string) (ws []string, es []error) {
	value := v.(string)
	count := utf8.RuneCountInString(value)
	if count < 1 {
		es = append(es, fmt.Errorf("%q cannot be less than 1 UTF-8 character", k))
	}

	if count > 32 {
		es = append(es, fmt.Errorf("%q cannot be longer than 32 UTF-8 characters", k))
	}
	if !regexache.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}\s]+`).MatchString(value) {
		es = append(es, fmt.Errorf(`%q must satisfy regular expression pattern: [\p{L}\p{M}\p{S}\p{N}\p{P}\s]+`, k))
	}
	return
}

func validResourceServerScopeName(v any, k string) (ws []string, errors []error) {
	value := v.(string)

	if len(value) < 1 {
		errors = append(errors, fmt.Errorf("%q cannot be less than 1 character", k))
	}
	if len(value) > 256 {
		errors = append(errors, fmt.Errorf("%q cannot be longer than 256 characters", k))
	}
	if !regexache.MustCompile(`[\x21\x23-\x2E\x30-\x5B\x5D-\x7E]+`).MatchString(value) {
		errors = append(errors, fmt.Errorf(`%q must satisfy regular expression pattern: [\x21\x23-\x2E\x30-\x5B\x5D-\x7E]+`, k))
	}
	return
}

func validUserGroupName(v any, k string) (ws []string, es []error) {
	value := v.(string)
	count := utf8.RuneCountInString(value)
	if count < 1 {
		es = append(es, fmt.Errorf("%q cannot be less than 1 UTF-8 character", k))
	}

	if count > 128 {
		es = append(es, fmt.Errorf("%q cannot be longer than 128 UTF-8 characters", k))
	}

	if !regexache.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}]+`).MatchString(value) {
		es = append(es, fmt.Errorf(`%q must satisfy regular expression pattern: [\p{L}\p{M}\p{S}\p{N}\p{P}]+`, k))
	}
	return
}

func validUserPoolEmailVerificationMessage(v any, k string) (ws []string, es []error) {
	value := v.(string)
	count := utf8.RuneCountInString(value)
	if count < 6 {
		es = append(es, fmt.Errorf("%q cannot be less than 6 UTF-8 characters", k))
	}

	if count > 20000 {
		es = append(es, fmt.Errorf("%q cannot be longer than 20000 UTF-8 characters", k))
	}

	if !regexache.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*\{####\}[\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*`).MatchString(value) {
		es = append(es, fmt.Errorf("%q does not contain {####}", k))
	}
	return
}

func validUserPoolEmailVerificationSubject(v any, k string) (ws []string, es []error) {
	value := v.(string)
	count := utf8.RuneCountInString(value)
	if count < 1 {
		es = append(es, fmt.Errorf("%q cannot be less than 1 UTF-8 character", k))
	}

	if count > 140 {
		es = append(es, fmt.Errorf("%q cannot be longer than 140 UTF-8 characters", k))
	}

	if !regexache.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}\s]+`).MatchString(value) {
		es = append(es, fmt.Errorf("%q can be composed of any kind of letter, symbols, numeric character, punctuation and whitespaces", k))
	}
	return
}

func validUserPoolID(v any, k string) (ws []string, es []error) {
	value := v.(string)
	if !regexache.MustCompile(`^[\w-]+_[0-9A-Za-z]+$`).MatchString(value) {
		es = append(es, fmt.Errorf("%q must be the region name followed by an underscore and then alphanumeric pattern", k))
	}
	return
}

func validUserPoolInviteTemplateEmailMessage(v any, k string) (ws []string, es []error) {
	value := v.(string)
	count := utf8.RuneCountInString(value)
	if count < 6 {
		es = append(es, fmt.Errorf("%q cannot be less than 6 UTF-8 characters", k))
	}

	if count > 20000 {
		es = append(es, fmt.Errorf("%q cannot be longer than 20000 UTF-8 characters", k))
	}

	if !regexache.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*\{####\}[\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*`).MatchString(value) {
		es = append(es, fmt.Errorf("%q does not contain {####}", k))
	}

	if !regexache.MustCompile(`.*\{username\}.*`).MatchString(value) {
		es = append(es, fmt.Errorf("%q does not contain {username}", k))
	}
	return
}

func validUserPoolInviteTemplateSMSMessage(v any, k string) (ws []string, es []error) {
	value := v.(string)
	count := utf8.RuneCountInString(value)
	if count < 6 {
		es = append(es, fmt.Errorf("%q cannot be less than 6 UTF-8 characters", k))
	}

	if count > 140 {
		es = append(es, fmt.Errorf("%q cannot be longer than 140 UTF-8 characters", k))
	}

	if !regexache.MustCompile(`.*\{####\}.*`).MatchString(value) {
		es = append(es, fmt.Errorf("%q does not contain {####}", k))
	}

	if !regexache.MustCompile(`.*\{username\}.*`).MatchString(value) {
		es = append(es, fmt.Errorf("%q does not contain {username}", k))
	}
	return
}

func validUserPoolSchemaName(v any, k string) (ws []string, es []error) {
	value := v.(string)
	if len(value) < 1 {
		es = append(es, fmt.Errorf("%q cannot be less than 1 character", k))
	}

	if len(value) > 20 {
		es = append(es, fmt.Errorf("%q cannot be longer than 20 characters", k))
	}

	if !regexache.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}]+`).MatchString(value) {
		es = append(es, fmt.Errorf(`%q must satisfy regular expression pattern: [\p{L}\p{M}\p{S}\p{N}\p{P}]+`, k))
	}
	return
}

func validUserPoolSMSAuthenticationMessage(v any, k string) (ws []string, es []error) {
	value := v.(string)
	count := utf8.RuneCountInString(value)
	if count < 6 {
		es = append(es, fmt.Errorf("%q cannot be less than 6 UTF-8 characters", k))
	}

	if count > 140 {
		es = append(es, fmt.Errorf("%q cannot be longer than 140 UTF-8 characters", k))
	}

	if !regexache.MustCompile(`.*\{####\}.*`).MatchString(value) {
		es = append(es, fmt.Errorf("%q does not contain {####}", k))
	}
	return
}

func validUserPoolSMSVerificationMessage(v any, k string) (ws []string, es []error) {
	value := v.(string)
	count := utf8.RuneCountInString(value)
	if count < 6 {
		es = append(es, fmt.Errorf("%q cannot be less than 6 UTF-8 characters", k))
	}

	if count > 140 {
		es = append(es, fmt.Errorf("%q cannot be longer than 140 UTF-8 characters", k))
	}

	if !regexache.MustCompile(`.*\{####\}.*`).MatchString(value) {
		es = append(es, fmt.Errorf("%q does not contain {####}", k))
	}
	return
}

func validUserPoolTemplateEmailMessage(v any, k string) (ws []string, es []error) {
	value := v.(string)
	count := utf8.RuneCountInString(value)
	if count < 6 {
		es = append(es, fmt.Errorf("%q cannot be less than 6 UTF-8 characters", k))
	}

	if count > 20000 {
		es = append(es, fmt.Errorf("%q cannot be longer than 20000 UTF-8 characters", k))
	}

	if !regexache.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*\{####\}[\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*`).MatchString(value) {
		es = append(es, fmt.Errorf("%q does not contain {####}", k))
	}
	return
}

func validUserPoolTemplateEmailMessageByLink(v any, k string) (ws []string, es []error) {
	value := v.(string)
	count := utf8.RuneCountInString(value)
	if count < 6 {
		es = append(es, fmt.Errorf("%q cannot be less than 6 UTF-8 characters", k))
	}

	if count > 20000 {
		es = append(es, fmt.Errorf("%q cannot be longer than 20000 UTF-8 characters", k))
	}

	if !regexache.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*\{##[\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*##\}[\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*`).MatchString(value) {
		es = append(es, fmt.Errorf(`%q must satisfy regular expression pattern: [\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*\{##[\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*##\}[\p{L}\p{M}\p{S}\p{N}\p{P}\s*]*`, k))
	}
	return
}

func validUserPoolTemplateEmailSubject(v any, k string) (ws []string, es []error) {
	value := v.(string)
	count := utf8.RuneCountInString(value)
	if count < 1 {
		es = append(es, fmt.Errorf("%q cannot be less than 1 UTF-8 character", k))
	}

	if count > 140 {
		es = append(es, fmt.Errorf("%q cannot be longer than 140 UTF-8 characters", k))
	}

	if !regexache.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}\s]+`).MatchString(value) {
		es = append(es, fmt.Errorf(`%q must satisfy regular expression pattern: [\p{L}\p{M}\p{S}\p{N}\p{P}\s]+`, k))
	}
	return
}

func validUserPoolTemplateEmailSubjectByLink(v any, k string) (ws []string, es []error) {
	value := v.(string)
	count := utf8.RuneCountInString(value)
	if count < 1 {
		es = append(es, fmt.Errorf("%q cannot be less than 1 UTF-8 character", k))
	}

	if count > 140 {
		es = append(es, fmt.Errorf("%q cannot be longer than 140 UTF-8 characters", k))
	}

	if !regexache.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}\s]+`).MatchString(value) {
		es = append(es, fmt.Errorf(`%q must satisfy regular expression pattern: [\p{L}\p{M}\p{S}\p{N}\p{P}\s]+`, k))
	}
	return
}

func validUserPoolTemplateSMSMessage(v any, k string) (ws []string, es []error) {
	value := v.(string)
	count := utf8.RuneCountInString(value)
	if count < 6 {
		es = append(es, fmt.Errorf("%q cannot be less than 6 UTF-8 characters", k))
	}

	if count > 140 {
		es = append(es, fmt.Errorf("%q cannot be longer than 140 UTF-8 characters", k))
	}

	if !regexache.MustCompile(`.*\{####\}.*`).MatchString(value) {
		es = append(es, fmt.Errorf("%q does not contain {####}", k))
	}
	return
}

var userPoolClientIdentityProviderValidator = []validator.String{
	stringvalidator.UTF8LengthBetween(1, 32),
	stringValidatorpLpMpSpNpP,
}

var userPoolClientNameValidator = []validator.String{
	stringvalidator.LengthBetween(1, 128),
	stringValidatorUserPoolClientName,
}

var userPoolClientURLValidator = []validator.String{
	stringvalidator.UTF8LengthBetween(1, 1024),
	stringValidatorpLpMpSpNpP,
}

var stringValidatorUserPoolClientName = stringvalidator.RegexMatches(
	regexache.MustCompile(`[\w\s+=,.@-]+`),
	`can include any letter, number, space, tab, or one of "+=,.@-"`,
)

var stringValidatorpLpMpSpNpP = stringvalidator.RegexMatches(
	regexache.MustCompile(`[\p{L}\p{M}\p{S}\p{N}\p{P}]+`),
	"can include any valid Unicode letter, combining character, symbol, number, or punctuation",
)
