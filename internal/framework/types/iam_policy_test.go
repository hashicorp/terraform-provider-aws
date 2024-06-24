// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestIAMPolicyValidateAttribute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         fwtypes.IAMPolicy
		expectError bool
	}
	tests := map[string]testCase{
		"unknown": {
			val: fwtypes.IAMPolicyUnknown(),
		},
		"null": {
			val: fwtypes.IAMPolicyNull(),
		},
		"valid": {
			val: fwtypes.IAMPolicyValue(`{"Key1": "Value", "Key2": [1, 2, 3]}`),
		},
		"invalid": {
			val:         fwtypes.IAMPolicyValue("not ok"),
			expectError: true,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			req := xattr.ValidateAttributeRequest{}
			resp := xattr.ValidateAttributeResponse{}

			test.val.ValidateAttribute(ctx, req, &resp)
			if resp.Diagnostics.HasError() != test.expectError {
				t.Errorf("resp.Diagnostics.HasError() = %t, want = %t", resp.Diagnostics.HasError(), test.expectError)
			}
		})
	}
}

func TestIAMPolicyStringSemanticEquals(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val1, val2 fwtypes.IAMPolicy
		equals     bool
	}
	tests := map[string]testCase{
		"both empty": {
			val1:   fwtypes.IAMPolicyValue(` `),
			val2:   fwtypes.IAMPolicyValue(`{}`),
			equals: true,
		},
		"not equals": {
			val1: fwtypes.IAMPolicyValue(`
{
  "Version": "2012-10-17",
  "Id": "S3AccessGrantsPolicy",
  "Statement": [{
    "Sid": "AllowAccessToS3AccessGrants",
    "Effect": "Allow",
    "Principal": {
      "AWS": "123456789456"
    },
    "Action": [
      "s3:ListAccessGrants",
      "s3:ListAccessGrantsLocations",
      "s3:GetDataAccess"
    ],
    "Resource": "arn:aws:s3:us-east-2:123456789123:access-grants/default"
  }]
}
`),
			val2: fwtypes.IAMPolicyValue(`
{
  "Version": "2012-10-17",
  "Id": "S3AccessGrantsPolicy",
  "Statement": [{
    "Sid": "AllowAccessToS3AccessGrants",
    "Effect": "Allow",
    "Resource": "arn:aws:s3:us-east-1:234567890123:access-grants/default"
    "Principal": {
      "AWS": "123456789456"
    },
    "Action": [
      "s3:ListAccessGrants",
      "s3:GetDataAccess"
    ]
  }]
}
`),
		},
		"equals": {
			val1: fwtypes.IAMPolicyValue(`
{
  "Version": "2012-10-17",
  "Id": "S3AccessGrantsPolicy",
  "Statement": [{
    "Sid": "AllowAccessToS3AccessGrants",
    "Effect": "Allow",
    "Principal": {
      "AWS": "123456789456"
    },
    "Action": [
      "s3:ListAccessGrants",
      "s3:ListAccessGrantsLocations",
      "s3:GetDataAccess"
    ],
    "Resource": "arn:aws:s3:us-east-2:123456789123:access-grants/default"
  }]
}
`),
			val2: fwtypes.IAMPolicyValue(`
{
  "Version": "2012-10-17",
  "Id": "S3AccessGrantsPolicy",
  "Statement": [{
    "Sid": "AllowAccessToS3AccessGrants",
    "Effect": "Allow",
    "Resource": "arn:aws:s3:us-east-2:123456789123:access-grants/default",
    "Principal": {
      "AWS": "arn:aws:iam::123456789456:root"
    },
    "Action": [
      "s3:ListAccessGrantsLocations",
      "s3:ListAccessGrants",
      "s3:GetDataAccess"
    ]
  }]
}
`),
			equals: true,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			equals, _ := test.val1.StringSemanticEquals(ctx, test.val2)

			if got, want := equals, test.equals; got != want {
				t.Errorf("StringSemanticEquals(%q, %q) = %v, want %v", test.val1, test.val2, got, want)
			}
		})
	}
}
