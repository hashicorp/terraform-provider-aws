package cognitoidp

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

func TestUserPoolSchemaAttributeMatchesStandardAttribute(t *testing.T) {
	cases := []struct {
		Input    *cognitoidentityprovider.SchemaAttributeType
		Expected bool
	}{
		{
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
		output := UserPoolSchemaAttributeMatchesStandardAttribute(tc.Input)
		if output != tc.Expected {
			t.Fatalf("Expected %t match with standard attribute on input: \n\n%#v\n\n", tc.Expected, tc.Input)
		}
	}
}
