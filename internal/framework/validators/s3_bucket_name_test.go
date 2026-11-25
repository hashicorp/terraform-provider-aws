// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package validators_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
)

func TestS3BucketNameValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         types.String
		expectError bool
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
			val:         types.StringValue("a"),
			expectError: true,
		},
		"invalid too short (2 chars)": {
			val:         types.StringValue("ab"),
			expectError: true,
		},
		"invalid too long (64 chars)": {
			val:         types.StringValue("a123456789012345678901234567890123456789012345678901234567890123"),
			expectError: true,
		},
		"invalid uppercase letters": {
			val:         types.StringValue("MyBucket"),
			expectError: true,
		},
		"invalid starts with hyphen": {
			val:         types.StringValue("-mybucket"),
			expectError: true,
		},
		"invalid ends with hyphen": {
			val:         types.StringValue("mybucket-"),
			expectError: true,
		},
		"invalid starts with period": {
			val:         types.StringValue(".mybucket"),
			expectError: true,
		},
		"invalid ends with period": {
			val:         types.StringValue("mybucket."),
			expectError: true,
		},
		"invalid contains underscore": {
			val:         types.StringValue("my_bucket"),
			expectError: true,
		},
		"invalid contains space": {
			val:         types.StringValue("my bucket"),
			expectError: true,
		},
		"invalid contains special characters": {
			val:         types.StringValue("my#bucket"),
			expectError: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()

			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			response := validator.StringResponse{}
			fwvalidators.S3BucketName.ValidateString(ctx, request, &response)

			if got, want := response.Diagnostics.HasError(), test.expectError; got != want {
				t.Errorf("S3BucketName.ValidateString() HasError() = %t, want = %t", got, want)
			}
		})
	}
}
