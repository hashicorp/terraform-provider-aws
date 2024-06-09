// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestDiffSuppressKeyID(t *testing.T) {
	t.Parallel()

	testcases := map[string]struct {
		old            string
		new            string
		expectSuppress bool
	}{
		"same ids": {
			old:            "57ff7a43-341d-46b6-aee3-a450c9de6dc8",
			new:            "57ff7a43-341d-46b6-aee3-a450c9de6dc8",
			expectSuppress: true,
		},
		"same arns": {
			old:            "arn:aws:kms:us-west-2:123456789012:key/57ff7a43-341d-46b6-aee3-a450c9de6dc8", // lintignore:AWSAT003,AWSAT005
			new:            "arn:aws:kms:us-west-2:123456789012:key/57ff7a43-341d-46b6-aee3-a450c9de6dc8", // lintignore:AWSAT003,AWSAT005
			expectSuppress: true,
		},

		"different ids": {
			old:            "57ff7a43-341d-46b6-aee3-a450c9de6dc8",
			new:            "2fc6ef7b-7a71-4426-ad52-0840169767f1",
			expectSuppress: false,
		},
		"different arns": {
			old:            "arn:aws:kms:us-west-2:123456789012:key/57ff7a43-341d-46b6-aee3-a450c9de6dc8", // lintignore:AWSAT003,AWSAT005
			new:            "arn:aws:kms:us-west-2:123456789012:key/2fc6ef7b-7a71-4426-ad52-0840169767f1", // lintignore:AWSAT003,AWSAT005
			expectSuppress: false,
		},

		"id to equivalent arn": {
			old:            "57ff7a43-341d-46b6-aee3-a450c9de6dc8",
			new:            "arn:aws:kms:us-west-2:123456789012:key/57ff7a43-341d-46b6-aee3-a450c9de6dc8", // lintignore:AWSAT003,AWSAT005
			expectSuppress: true,
		},
		"arn to equivalent id": {
			old:            "arn:aws:kms:us-west-2:123456789012:key/57ff7a43-341d-46b6-aee3-a450c9de6dc8", // lintignore:AWSAT003,AWSAT005
			new:            "57ff7a43-341d-46b6-aee3-a450c9de6dc8",
			expectSuppress: true,
		},

		"id to different arn": {
			old:            "57ff7a43-341d-46b6-aee3-a450c9de6dc8",
			new:            "arn:aws:kms:us-west-2:123456789012:key/2fc6ef7b-7a71-4426-ad52-0840169767f1", // lintignore:AWSAT003,AWSAT005
			expectSuppress: false,
		},
		"arn to different id": {
			old:            "arn:aws:kms:us-west-2:123456789012:key/57ff7a43-341d-46b6-aee3-a450c9de6dc8", // lintignore:AWSAT003,AWSAT005
			new:            "2fc6ef7b-7a71-4426-ad52-0840169767f1",
			expectSuppress: false,
		},
	}

	for name, testcase := range testcases {
		testcase := testcase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := diffSuppressKey(names.AttrField, testcase.old, testcase.new, nil)

			if e := testcase.expectSuppress; actual != e {
				t.Fatalf("expected %t, got %t", e, actual)
			}
		})
	}
}
