// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validators_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
)

func TestARNValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val                 types.String
		expectedDiagnostics diag.Diagnostics
	}
	tests := map[string]testCase{
		"unknown String": {
			val: types.StringUnknown(),
		},
		"null String": {
			val: types.StringNull(),
		},
		"Beanstalk": {
			val: types.StringValue("arn:aws:elasticbeanstalk:us-east-1:123456789012:environment/My App/MyEnvironment"),
		},
		"IAM User": {
			val: types.StringValue("arn:aws:iam::123456789012:user/David"),
		},
		"Managed IAM policy": {
			val: types.StringValue("arn:aws:iam::aws:policy/CloudWatchReadOnlyAccess"),
		},
		"ImageBuilder": {
			val: types.StringValue("arn:aws:imagebuilder:us-east-1:third-party:component/my-component"),
		},
		"RDS": {
			val: types.StringValue("arn:aws:rds:eu-west-1:123456789012:db:mysql-db"),
		},
		"S3 object": {
			val: types.StringValue("arn:aws:s3:::my_corporate_bucket/exampleobject.png"),
		},
		"CloudWatch Rule": {
			val: types.StringValue("arn:aws:events:us-east-1:319201112229:rule/rule_name"),
		},
		"Lambda function": {
			val: types.StringValue("arn:aws:lambda:eu-west-1:319201112229:function:myCustomFunction"),
		},
		"Lambda func qualifier": {
			val: types.StringValue("arn:aws:lambda:eu-west-1:319201112229:function:myCustomFunction:Qualifier"),
		},
		"China EC2 ARN": {
			val: types.StringValue("arn:aws-cn:ec2:cn-north-1:123456789012:instance/i-12345678"),
		},
		"China S3 ARN": {
			val: types.StringValue("arn:aws-cn:s3:::bucket/object"),
		},
		"C2S EC2 ARN": {
			val: types.StringValue("arn:aws-iso:ec2:us-iso-east-1:123456789012:instance/i-12345678"),
		},
		"C2S S3 ARN": {
			val: types.StringValue("arn:aws-iso:s3:::bucket/object"),
		},
		"SC2S EC2 ARN": {
			val: types.StringValue("arn:aws-iso-b:ec2:us-isob-east-1:123456789012:instance/i-12345678"),
		},
		"SC2S S3 ARN": {
			val: types.StringValue("arn:aws-iso-b:s3:::bucket/object"),
		},
		"GovCloud EC2 ARN": {
			val: types.StringValue("arn:aws-us-gov:ec2:us-gov-west-1:123456789012:instance/i-12345678"),
		},
		"GovCloud S3 ARN": {
			val: types.StringValue("arn:aws-us-gov:s3:::bucket/object"),
		},
		"Cloudwatch Alarm": {
			val: types.StringValue("arn:aws:cloudwatch::cw0000000000:alarm:my-alarm"),
		},
		"Invalid prefix 1": {
			val: types.StringValue("arn"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid arn, got: (arn) is an invalid ARN: arn: invalid prefix`,
				),
			},
		},
		"Invalid prefix 2": {
			val: types.StringValue("123456789012"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid arn, got: (123456789012) is an invalid ARN: arn: invalid prefix`,
				),
			},
		},
		"Not enough sections 1": {
			val: types.StringValue("arn:aws"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid arn, got: (arn:aws) is an invalid ARN: arn: not enough sections`,
				),
			},
		},
		"Not enough sections 2": {
			val: types.StringValue("arn:aws:logs"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid arn, got: (arn:aws:logs) is an invalid ARN: arn: not enough sections`,
				),
			},
		},
		"Invalid region value": {
			val: types.StringValue("arn:aws:logs:region:*:*"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid arn, got: (arn:aws:logs:region:*:*) is an invalid ARN: invalid region value (expecting to match regular expression: ^[a-z]{2}(-[a-z]+)+-\d$)`,
				),
			},
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			response := validator.StringResponse{}
			fwvalidators.ARN().ValidateString(ctx, request, &response)

			if diff := cmp.Diff(response.Diagnostics, test.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
