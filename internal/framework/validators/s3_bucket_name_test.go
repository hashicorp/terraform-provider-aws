// Copyright IBM Corp. 2014, 2026
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

func TestS3BucketNameValidator(t *testing.T) {
	t.Parallel()

	errorSummary := "Invalid Attribute Value Match"
	errorDetail := "Attribute test value must match regular expression: " +
		`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$` +
		"\nBucket names must be 3 to 63 characters and begin and end with a letter or number. " +
		"Valid characters are a-z, 0-9, periods (.), and hyphens."

	newError := func(got string) diag.Diagnostics {
		return diag.Diagnostics{
			diag.NewAttributeErrorDiagnostic(
				path.Root("test"),
				errorSummary,
				"Attribute test value must match regular expression: "+
					`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`+"\n"+
					"Bucket names must be 3 to 63 characters and begin and end with a letter or number. "+
					"Valid characters are a-z, 0-9, periods (.), and hyphens., got: "+got,
			),
		}
	}
	_ = errorDetail // suppress unused-variable lint if the helper above is used instead

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
		// Valid names
		"valid lowercase letters only": {
			val: types.StringValue("mybucket"),
		},
		"valid letters and numbers": {
			val: types.StringValue("my-bucket-123"),
		},
		"valid with periods": {
			val: types.StringValue("my.bucket.name"),
		},
		"valid minimum length (3 chars)": {
			val: types.StringValue("abc"),
		},
		"valid maximum length (63 chars)": {
			// 63 characters: starts and ends with letter, 61 middle chars
			val: types.StringValue("a" + "b-c.d-e.f-g.h-i.j-k.l-m.n-o.p-q.r-s.t-u.v-w.x-y.z-0.1-2.3" + "a"),
		},
		"valid starts with digit": {
			val: types.StringValue("0mybucket"),
		},
		"valid ends with digit": {
			val: types.StringValue("mybucket0"),
		},
		// Invalid names
		"invalid too short (1 char)": {
			val:                 types.StringValue("a"),
			expectedDiagnostics: newError("a"),
		},
		"invalid too short (2 chars)": {
			val:                 types.StringValue("ab"),
			expectedDiagnostics: newError("ab"),
		},
		"invalid too long (64 chars)": {
			val:                 types.StringValue("a123456789012345678901234567890123456789012345678901234567890123"),
			expectedDiagnostics: newError("a123456789012345678901234567890123456789012345678901234567890123"),
		},
		"invalid uppercase letters": {
			val:                 types.StringValue("MyBucket"),
			expectedDiagnostics: newError("MyBucket"),
		},
		"invalid starts with hyphen": {
			val:                 types.StringValue("-mybucket"),
			expectedDiagnostics: newError("-mybucket"),
		},
		"invalid ends with hyphen": {
			val:                 types.StringValue("mybucket-"),
			expectedDiagnostics: newError("mybucket-"),
		},
		"invalid starts with period": {
			val:                 types.StringValue(".mybucket"),
			expectedDiagnostics: newError(".mybucket"),
		},
		"invalid ends with period": {
			val:                 types.StringValue("mybucket."),
			expectedDiagnostics: newError("mybucket."),
		},
		"invalid contains underscore": {
			val:                 types.StringValue("my_bucket"),
			expectedDiagnostics: newError("my_bucket"),
		},
		"invalid contains space": {
			val:                 types.StringValue("my bucket"),
			expectedDiagnostics: newError("my bucket"),
		},
		"invalid contains special characters": {
			val:                 types.StringValue("my#bucket"),
			expectedDiagnostics: newError("my#bucket"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			response := validator.StringResponse{}
			fwvalidators.S3BucketName.ValidateString(ctx, request, &response)

			if diff := cmp.Diff(response.Diagnostics, test.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
