// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestUserPoolSchemaAttributeMatchesStandardAttribute(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name     string
		Input    *awstypes.SchemaAttributeType
		Expected bool
	}{
		{
			Name: "birthday standard",
			Input: &awstypes.SchemaAttributeType{
				AttributeDataType:      awstypes.AttributeDataTypeString,
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("birthdate"),
				Required:               aws.Bool(false),
				StringAttributeConstraints: &awstypes.StringAttributeConstraintsType{
					MaxLength: aws.String("10"), // nosemgrep:ci.literal-10-string-test-constant
					MinLength: aws.String("10"), // nosemgrep:ci.literal-10-string-test-constant
				},
			},
			Expected: true,
		},
		{
			Name: "birthday non-standard DeveloperOnlyAttribute",
			Input: &awstypes.SchemaAttributeType{
				AttributeDataType:      awstypes.AttributeDataTypeString,
				DeveloperOnlyAttribute: aws.Bool(true),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("birthdate"),
				Required:               aws.Bool(false),
				StringAttributeConstraints: &awstypes.StringAttributeConstraintsType{
					MaxLength: aws.String("10"), // nosemgrep:ci.literal-10-string-test-constant
					MinLength: aws.String("10"), // nosemgrep:ci.literal-10-string-test-constant
				},
			},
			Expected: false,
		},
		{
			Name: "birthday non-standard Mutable",
			Input: &awstypes.SchemaAttributeType{
				AttributeDataType:      awstypes.AttributeDataTypeString,
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(false),
				Name:                   aws.String("birthdate"),
				Required:               aws.Bool(false),
				StringAttributeConstraints: &awstypes.StringAttributeConstraintsType{
					MaxLength: aws.String("10"), // nosemgrep:ci.literal-10-string-test-constant
					MinLength: aws.String("10"), // nosemgrep:ci.literal-10-string-test-constant
				},
			},
			Expected: false,
		},
		{
			Name: "non-standard Name",
			Input: &awstypes.SchemaAttributeType{
				AttributeDataType:      awstypes.AttributeDataTypeString,
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("non-existent"),
				Required:               aws.Bool(false),
				StringAttributeConstraints: &awstypes.StringAttributeConstraintsType{
					MaxLength: aws.String("10"), // nosemgrep:ci.literal-10-string-test-constant
					MinLength: aws.String("10"), // nosemgrep:ci.literal-10-string-test-constant
				},
			},
			Expected: false,
		},
		{
			Name: "birthday non-standard Required",
			Input: &awstypes.SchemaAttributeType{
				AttributeDataType:      awstypes.AttributeDataTypeString,
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("birthdate"),
				Required:               aws.Bool(true),
				StringAttributeConstraints: &awstypes.StringAttributeConstraintsType{
					MaxLength: aws.String("10"), // nosemgrep:ci.literal-10-string-test-constant
					MinLength: aws.String("10"), // nosemgrep:ci.literal-10-string-test-constant
				},
			},
			Expected: false,
		},
		{
			Name: "birthday non-standard StringAttributeConstraints.Max",
			Input: &awstypes.SchemaAttributeType{
				AttributeDataType:      awstypes.AttributeDataTypeString,
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("birthdate"),
				Required:               aws.Bool(false),
				StringAttributeConstraints: &awstypes.StringAttributeConstraintsType{
					MaxLength: aws.String("999"),
					MinLength: aws.String("10"), // nosemgrep:ci.literal-10-string-test-constant
				},
			},
			Expected: false,
		},
		{
			Name: "birthday non-standard StringAttributeConstraints.Min",
			Input: &awstypes.SchemaAttributeType{
				AttributeDataType:      awstypes.AttributeDataTypeString,
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("birthdate"),
				Required:               aws.Bool(false),
				StringAttributeConstraints: &awstypes.StringAttributeConstraintsType{
					MaxLength: aws.String("10"), // nosemgrep:ci.literal-10-string-test-constant
					MinLength: aws.String("999"),
				},
			},
			Expected: false,
		},
		{
			Name: "email_verified standard",
			Input: &awstypes.SchemaAttributeType{
				AttributeDataType:      awstypes.AttributeDataTypeBoolean,
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("email_verified"),
				Required:               aws.Bool(false),
			},
			Expected: true,
		},
		{
			Name: "updated_at standard",
			Input: &awstypes.SchemaAttributeType{
				AttributeDataType:      awstypes.AttributeDataTypeNumber,
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("updated_at"),
				NumberAttributeConstraints: &awstypes.NumberAttributeConstraintsType{
					MinValue: aws.String("0"), // nosemgrep:ci.literal-0-string-test-constant
				},
				Required: aws.Bool(false),
			},
			Expected: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			output := userPoolSchemaAttributeMatchesStandardAttribute(tc.Input)
			if output != tc.Expected {
				t.Fatalf("Expected %t match with standard attribute on input: \n\n%#v\n\n", tc.Expected, tc.Input)
			}
		})
	}
}

func TestSkipFlatteningStringAttributeContraints(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		configured []awstypes.SchemaAttributeType
		input      *awstypes.SchemaAttributeType
		want       bool
	}{
		{
			name: "config omitted",
			configured: []awstypes.SchemaAttributeType{
				{
					AttributeDataType:      awstypes.AttributeDataTypeString,
					DeveloperOnlyAttribute: aws.Bool(false),
					Mutable:                aws.Bool(false),
					Name:                   aws.String(names.AttrEmail),
					Required:               aws.Bool(true),
				},
			},
			input: &awstypes.SchemaAttributeType{
				AttributeDataType:      awstypes.AttributeDataTypeString,
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(false),
				Name:                   aws.String(names.AttrEmail),
				Required:               aws.Bool(true),
				StringAttributeConstraints: &awstypes.StringAttributeConstraintsType{
					MaxLength: aws.String("2048"),
					MinLength: aws.String("0"), // nosemgrep:ci.literal-0-string-test-constant
				},
			},
			want: true,
		},
		{
			name: "config set",
			configured: []awstypes.SchemaAttributeType{
				{
					AttributeDataType:      awstypes.AttributeDataTypeString,
					DeveloperOnlyAttribute: aws.Bool(false),
					Mutable:                aws.Bool(false),
					Name:                   aws.String(names.AttrEmail),
					Required:               aws.Bool(true),
					StringAttributeConstraints: &awstypes.StringAttributeConstraintsType{
						MaxLength: aws.String("2048"),
						MinLength: aws.String("0"), // nosemgrep:ci.literal-0-string-test-constant
					},
				},
			},
			input: &awstypes.SchemaAttributeType{
				AttributeDataType:      awstypes.AttributeDataTypeString,
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(false),
				Name:                   aws.String(names.AttrEmail),
				Required:               aws.Bool(true),
				StringAttributeConstraints: &awstypes.StringAttributeConstraintsType{
					MaxLength: aws.String("2048"),
					MinLength: aws.String("0"), // nosemgrep:ci.literal-0-string-test-constant
				},
			},
			want: false,
		},
		{
			name: "config set with diff",
			configured: []awstypes.SchemaAttributeType{
				{
					AttributeDataType:      awstypes.AttributeDataTypeString,
					DeveloperOnlyAttribute: aws.Bool(false),
					Mutable:                aws.Bool(false),
					Name:                   aws.String(names.AttrEmail),
					Required:               aws.Bool(true),
					StringAttributeConstraints: &awstypes.StringAttributeConstraintsType{
						MaxLength: aws.String("1024"),
						MinLength: aws.String("5"),
					},
				},
			},
			input: &awstypes.SchemaAttributeType{
				AttributeDataType:      awstypes.AttributeDataTypeString,
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(false),
				Name:                   aws.String(names.AttrEmail),
				Required:               aws.Bool(true),
				StringAttributeConstraints: &awstypes.StringAttributeConstraintsType{
					MaxLength: aws.String("2048"),
					MinLength: aws.String("0"), // nosemgrep:ci.literal-0-string-test-constant
				},
			},
			want: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := skipFlatteningStringAttributeContraints(tc.configured, tc.input)
			if got != tc.want {
				t.Fatalf("skipFlatteningStringAttributeContraints() got %t, want %t\n\n%#v\n\n", got, tc.want, tc.input)
			}
		})
	}
}
