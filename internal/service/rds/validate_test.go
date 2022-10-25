package rds

import (
	"strings"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
)

func TestValidEventSubscriptionName(t *testing.T) {
	validNames := []string{
		"valid-name",
		"valid02-name",
		"Valid-Name1",
	}
	for _, v := range validNames {
		_, errors := validEventSubscriptionName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid RDS Event Subscription Name: %q", v, errors)
		}
	}

	invalidNames := []string{
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
		_, errors := validEventSubscriptionName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid RDS Event Subscription Name", v)
		}
	}
}

func TestValidOptionGroupName(t *testing.T) {
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
			Value:    sdkacctest.RandStringFromCharSet(256, sdkacctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validOptionGroupName(tc.Value, "aws_db_option_group_name")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the DB Option Group Name to trigger a validation error")
		}
	}
}

func TestValidOptionGroupNamePrefix(t *testing.T) {
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
			Value:    sdkacctest.RandStringFromCharSet(230, sdkacctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validOptionGroupNamePrefix(tc.Value, "aws_db_option_group_name")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the DB Option Group name prefix to trigger a validation error")
		}
	}
}

func TestValidParamGroupName(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
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
			Value:    sdkacctest.RandStringFromCharSet(256, sdkacctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validParamGroupName(tc.Value, "aws_db_parameter_group_name")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the DB Parameter Group Name to trigger a validation error")
		}
	}
}

func TestValidSubnetGroupName(t *testing.T) {
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
			Value:    sdkacctest.RandStringFromCharSet(300, sdkacctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validSubnetGroupName(tc.Value, "aws_db_subnet_group")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the DB Subnet Group name to trigger a validation error")
		}
	}
}

func TestValidSubnetGroupNamePrefix(t *testing.T) {
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
			Value:    sdkacctest.RandStringFromCharSet(230, sdkacctest.CharSetAlpha),
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validSubnetGroupNamePrefix(tc.Value, "aws_db_subnet_group")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected the DB Subnet Group name prefix to trigger a validation error")
		}
	}
}
