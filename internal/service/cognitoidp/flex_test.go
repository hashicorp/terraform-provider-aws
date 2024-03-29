// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

func TestUserPoolSchemaAttributeMatchesStandardAttribute(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Name     string
		Input    *cognitoidentityprovider.SchemaAttributeType
		Expected bool
	}{
		{
			Name: "birthday standard",
			Input: &cognitoidentityprovider.SchemaAttributeType{
				AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("birthdate"),
				Required:               aws.Bool(false),
				StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
					MaxLength: aws.String("10"),
					MinLength: aws.String("10"),
				},
			},
			Expected: true,
		},
		{
			Name: "birthday non-standard DeveloperOnlyAttribute",
			Input: &cognitoidentityprovider.SchemaAttributeType{
				AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
				DeveloperOnlyAttribute: aws.Bool(true),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("birthdate"),
				Required:               aws.Bool(false),
				StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
					MaxLength: aws.String("10"),
					MinLength: aws.String("10"),
				},
			},
			Expected: false,
		},
		{
			Name: "birthday non-standard Mutable",
			Input: &cognitoidentityprovider.SchemaAttributeType{
				AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(false),
				Name:                   aws.String("birthdate"),
				Required:               aws.Bool(false),
				StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
					MaxLength: aws.String("10"),
					MinLength: aws.String("10"),
				},
			},
			Expected: false,
		},
		{
			Name: "non-standard Name",
			Input: &cognitoidentityprovider.SchemaAttributeType{
				AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("non-existent"),
				Required:               aws.Bool(false),
				StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
					MaxLength: aws.String("10"),
					MinLength: aws.String("10"),
				},
			},
			Expected: false,
		},
		{
			Name: "birthday non-standard Required",
			Input: &cognitoidentityprovider.SchemaAttributeType{
				AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("birthdate"),
				Required:               aws.Bool(true),
				StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
					MaxLength: aws.String("10"),
					MinLength: aws.String("10"),
				},
			},
			Expected: false,
		},
		{
			Name: "birthday non-standard StringAttributeConstraints.Max",
			Input: &cognitoidentityprovider.SchemaAttributeType{
				AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("birthdate"),
				Required:               aws.Bool(false),
				StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
					MaxLength: aws.String("999"),
					MinLength: aws.String("10"),
				},
			},
			Expected: false,
		},
		{
			Name: "birthday non-standard StringAttributeConstraints.Min",
			Input: &cognitoidentityprovider.SchemaAttributeType{
				AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("birthdate"),
				Required:               aws.Bool(false),
				StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
					MaxLength: aws.String("10"),
					MinLength: aws.String("999"),
				},
			},
			Expected: false,
		},
		{
			Name: "email_verified standard",
			Input: &cognitoidentityprovider.SchemaAttributeType{
				AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeBoolean),
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("email_verified"),
				Required:               aws.Bool(false),
			},
			Expected: true,
		},
		{
			Name: "updated_at standard",
			Input: &cognitoidentityprovider.SchemaAttributeType{
				AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeNumber),
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(true),
				Name:                   aws.String("updated_at"),
				NumberAttributeConstraints: &cognitoidentityprovider.NumberAttributeConstraintsType{
					MinValue: aws.String("0"),
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
			output := UserPoolSchemaAttributeMatchesStandardAttribute(tc.Input)
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
		configured []*cognitoidentityprovider.SchemaAttributeType
		input      *cognitoidentityprovider.SchemaAttributeType
		want       bool
	}{
		{
			name: "config omitted",
			configured: []*cognitoidentityprovider.SchemaAttributeType{
				{
					AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
					DeveloperOnlyAttribute: aws.Bool(false),
					Mutable:                aws.Bool(false),
					Name:                   aws.String("email"),
					Required:               aws.Bool(true),
				},
			},
			input: &cognitoidentityprovider.SchemaAttributeType{
				AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(false),
				Name:                   aws.String("email"),
				Required:               aws.Bool(true),
				StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
					MaxLength: aws.String("2048"),
					MinLength: aws.String("0"),
				},
			},
			want: true,
		},
		{
			name: "config set",
			configured: []*cognitoidentityprovider.SchemaAttributeType{
				{
					AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
					DeveloperOnlyAttribute: aws.Bool(false),
					Mutable:                aws.Bool(false),
					Name:                   aws.String("email"),
					Required:               aws.Bool(true),
					StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
						MaxLength: aws.String("2048"),
						MinLength: aws.String("0"),
					},
				},
			},
			input: &cognitoidentityprovider.SchemaAttributeType{
				AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(false),
				Name:                   aws.String("email"),
				Required:               aws.Bool(true),
				StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
					MaxLength: aws.String("2048"),
					MinLength: aws.String("0"),
				},
			},
			want: false,
		},
		{
			name: "config set with diff",
			configured: []*cognitoidentityprovider.SchemaAttributeType{
				{
					AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
					DeveloperOnlyAttribute: aws.Bool(false),
					Mutable:                aws.Bool(false),
					Name:                   aws.String("email"),
					Required:               aws.Bool(true),
					StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
						MaxLength: aws.String("1024"),
						MinLength: aws.String("5"),
					},
				},
			},
			input: &cognitoidentityprovider.SchemaAttributeType{
				AttributeDataType:      aws.String(cognitoidentityprovider.AttributeDataTypeString),
				DeveloperOnlyAttribute: aws.Bool(false),
				Mutable:                aws.Bool(false),
				Name:                   aws.String("email"),
				Required:               aws.Bool(true),
				StringAttributeConstraints: &cognitoidentityprovider.StringAttributeConstraintsType{
					MaxLength: aws.String("2048"),
					MinLength: aws.String("0"),
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
