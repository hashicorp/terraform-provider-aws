// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestValidUserGroupName(t *testing.T) {
	t.Parallel()

	validValues := []string{
		"foo",
		"7346241598935552",
		"foo_bar",
		"foo:bar",
		"foo/bar",
		"foo-bar",
		"$foobar",
		strings.Repeat("W", 128),
	}

	for _, s := range validValues {
		_, errors := validUserGroupName(s, names.AttrName)
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito User Pool Group Name: %v", s, errors)
		}
	}

	invalidValues := []string{
		"",
		strings.Repeat("W", 129), // > 128
	}

	for _, s := range invalidValues {
		_, errors := validUserGroupName(s, names.AttrName)
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito User Pool Group Name: %v", s, errors)
		}
	}
}

func TestValidUserPoolEmailVerificationMessage(t *testing.T) {
	t.Parallel()

	validValues := []string{
		"{####}",
		"Foo {####}",
		"{####} Bar",
		"AZERTYUIOPQSDFGHJKLMWXCVBN?./+%£*¨°0987654321&é\"'(§è!çà)-@^'{####},=ù`$|´”’[å»ÛÁØ]–Ô¥#‰±•",
		"{####}" + strings.Repeat("W", 19994), // = 20000
		"{####}" + strings.Repeat("あ", 19994), // = 20000, UTF-8 (2 bytes/char)
	}

	for _, s := range validValues {
		_, errors := validUserPoolEmailVerificationMessage(s, "email_verification_message")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito User Pool email verification message: %v", s, errors)
		}
	}

	invalidValues := []string{
		"Foo",
		"あいうえお",
		"{###}",
		"{####}" + strings.Repeat("W", 19995), // > 20000
		"{####}" + strings.Repeat("あ", 19995), // > 20000, UTF-8 (2 bytes/char)
	}

	for _, s := range invalidValues {
		_, errors := validUserPoolEmailVerificationMessage(s, "email_verification_message")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito User Pool email verification message: %v", s, errors)
		}
	}
}

func TestValidUserPoolEmailVerificationSubject(t *testing.T) {
	t.Parallel()

	validValues := []string{
		"FooBar",
		"AZERTYUIOPQSDFGHJKLMWXCVBN?./+%£*¨°0987654321&é\" '(§è!çà)-@^'{####},=ù`$|´”’[å»ÛÁØ]–Ô¥#‰±•",
		"Foo Bar", // special whitespace character
		strings.Repeat("W", 140),
		strings.Repeat("あ", 140), // UTF-8 (2 bytes/char)
	}

	for _, s := range validValues {
		_, errors := validUserPoolEmailVerificationSubject(s, "email_verification_subject")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito User Pool email verification subject: %v", s, errors)
		}
	}

	invalidValues := []string{
		strings.Repeat("W", 141), // > 140
		strings.Repeat("あ", 141), // UTF-8 (2 bytes/char)
	}

	for _, s := range invalidValues {
		_, errors := validUserPoolEmailVerificationSubject(s, "email_verification_subject")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito User Pool email verification subject: %v", s, errors)
		}
	}
}

func TestValidUserPoolID(t *testing.T) {
	t.Parallel()

	validValues := []string{
		"eu-west-1_Foo123",         //lintignore:AWSAT003
		"ap-southeast-2_BaRBaz987", //lintignore:AWSAT003
	}

	for _, s := range validValues {
		_, errors := validUserPoolID(s, names.AttrUserPoolID)
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito User Pool Id: %v", s, errors)
		}
	}

	invalidValues := []string{
		"",
		"foo",
		"us-east-1-Foo123",   //lintignore:AWSAT003
		"eu-central-2_Bar+4", //lintignore:AWSAT003
	}

	for _, s := range invalidValues {
		_, errors := validUserPoolID(s, names.AttrUserPoolID)
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito User Pool Id: %v", s, errors)
		}
	}
}

func TestValidUserPoolSMSAuthenticationMessage(t *testing.T) {
	t.Parallel()

	validValues := []string{
		"{####}",
		"Foo {####}",
		"{####} Bar",
		"AZERTYUIOPQSDFGHJKLMWXCVBN?./+%£*¨°0987654321&é\"'(§è!çà)-@^'{####},=ù`$|´”’[å»ÛÁØ]–Ô¥#‰±•",
		"{####}" + strings.Repeat("W", 134), // = 140
		"{####}" + strings.Repeat("あ", 134), // = 140, UTF-8 (2 bytes/char)
	}

	for _, s := range validValues {
		_, errors := validUserPoolSMSAuthenticationMessage(s, "sms_authentication_message")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito User Pool sms authentication message: %v", s, errors)
		}
	}

	invalidValues := []string{
		"Foo",
		"あいうえお",
		"{####}" + strings.Repeat("W", 135), // > 140
		"{####}" + strings.Repeat("あ", 135), // > 140, UTF-8 (2 bytes/char)
	}

	for _, s := range invalidValues {
		_, errors := validUserPoolSMSAuthenticationMessage(s, "sms_authentication_message")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito User Pool sms authentication message: %v", s, errors)
		}
	}
}

func TestValidUserPoolSMSVerificationMessage(t *testing.T) {
	t.Parallel()

	validValues := []string{
		"{####}",
		"Foo {####}",
		"{####} Bar",
		"AZERTYUIOPQSDFGHJKLMWXCVBN?./+%£*¨°0987654321&é\"'(§è!çà)-@^'{####},=ù`$|´”’[å»ÛÁØ]–Ô¥#‰±•",
		"{####}" + strings.Repeat("W", 134), // = 140
		"{####}" + strings.Repeat("あ", 134), // = 140, UTF-8 (2 bytes/char)
	}

	for _, s := range validValues {
		_, errors := validUserPoolSMSVerificationMessage(s, "sms_verification_message")
		if len(errors) > 0 {
			t.Fatalf("%q should be a valid Cognito User Pool sms authentication message: %v", s, errors)
		}
	}

	invalidValues := []string{
		"Foo",
		"あいうえお",
		"{####}" + strings.Repeat("W", 135), // > 140
		"{####}" + strings.Repeat("あ", 135), // > 140, UTF-8 (2 bytes/char)
	}

	for _, s := range invalidValues {
		_, errors := validUserPoolSMSVerificationMessage(s, "sms_verification_message")
		if len(errors) == 0 {
			t.Fatalf("%q should not be a valid Cognito User Pool sms authentication message: %v", s, errors)
		}
	}
}
