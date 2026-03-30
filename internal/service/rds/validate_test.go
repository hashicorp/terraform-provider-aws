// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfrds "github.com/hashicorp/terraform-provider-aws/internal/service/rds"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestValidEventSubscriptionName(t *testing.T) {
	t.Parallel()

	validNames := []string{
		"valid-name",
		"valid02-name",
		"Valid-Name1",
	}
	for _, v := range validNames {
		_, errors := tfrds.ValidEventSubscriptionName(v, names.AttrName)
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid RDS Event Subscription Name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"invalid_name",
		"invalid-name-",
		"-invalid-name",
		"0invalid-name",
		"invalid--name",
		"Here is a name with: colon",
		"and here is another * invalid name",
		"also $ invalid",
		"This . is also %% invalid@!)+(",
		"*",
		"",
		" ",
		"_",
		// length > 255
		strings.Repeat("W", 256),
	}
	for _, v := range invalidNames {
		_, errors := tfrds.ValidEventSubscriptionName(v, names.AttrName)
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid RDS Event Subscription Name", v)
		}
	}
}

func TestValidEventSubscriptionNamePrefix(t *testing.T) {
	t.Parallel()

	validNames := []string{
		"valid-name",
		"valid02-name",
		"Valid-Name1",
		"valid-name-",
	}
	for _, v := range validNames {
		_, errors := tfrds.ValidEventSubscriptionNamePrefix(v, names.AttrNamePrefix)
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid RDS Event Subscription Name: %q", v, errors)
		}
	}

	invalidNames := []string{
		"invalid_name",
		"-invalid-name",
		"0invalid-name",
		"invalid--name",
		"Here is a name with: colon",
		"and here is another * invalid name",
		"also $ invalid",
		"This . is also %% invalid@!)+(",
		"*",
		"",
		" ",
		"_",
		// length > 229
		strings.Repeat("W", 230),
	}
	for _, v := range invalidNames {
		_, errors := tfrds.ValidEventSubscriptionNamePrefix(v, names.AttrNamePrefix)
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid RDS Event Subscription Name Prefix", v)
		}
	}
}

func TestValidOptionGroupName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "testing123!",
			ErrCount: 1,
		},
		{
			Value:    "1testing123",
			ErrCount: 1,
		},
		{
			Value:    "testing--123",
			ErrCount: 1,
		},
		{
			Value:    "testing123-",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(t, 256, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := tfrds.ValidOptionGroupName(tc.Value, "aws_db_option_group_name")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the DB Option Group Name to trigger a validation error")
		}
	}
}

func TestValidOptionGroupNamePrefix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "testing123!",
			ErrCount: 1,
		},
		{
			Value:    "1testing123",
			ErrCount: 1,
		},
		{
			Value:    "testing--123",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(t, 230, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := tfrds.ValidOptionGroupNamePrefix(tc.Value, "aws_db_option_group_name")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the DB Option Group name prefix to trigger a validation error")
		}
	}
}

func TestValidParamGroupName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "default.postgres9.6",
			ErrCount: 0,
		},
		{
			Value:    "tEsting123",
			ErrCount: 1,
		},
		{
			Value:    "testing123!",
			ErrCount: 1,
		},
		{
			Value:    "1testing123",
			ErrCount: 1,
		},
		{
			Value:    "testing--123",
			ErrCount: 1,
		},
		{
			Value:    "testing_123",
			ErrCount: 1,
		},
		{
			Value:    "testing123-",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(t, 256, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := tfrds.ValidParamGroupName(tc.Value, "aws_db_parameter_group_name")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the DB Parameter Group Name to trigger a validation error")
		}
	}
}

func TestValidSubnetGroupName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "tEsting",
			ErrCount: 1,
		},
		{
			Value:    "testing?",
			ErrCount: 1,
		},
		{
			Value:    "default",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(t, 300, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := tfrds.ValidSubnetGroupName(tc.Value, "aws_db_subnet_group")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the DB Subnet Group name to trigger a validation error")
		}
	}
}

func TestValidSubnetGroupNamePrefix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "tEsting",
			ErrCount: 1,
		},
		{
			Value:    "testing?",
			ErrCount: 1,
		},
		{
			Value:    acctest.RandStringFromCharSet(t, 230, acctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := tfrds.ValidSubnetGroupNamePrefix(tc.Value, "aws_db_subnet_group")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the DB Subnet Group name prefix to trigger a validation error")
		}
	}
}
