// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package firehose_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tffirehose "github.com/hashicorp/terraform-provider-aws/internal/service/firehose"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestMigrateState(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
		Meta         interface{}
	}{
		"v0.6.16 and earlier": {
			StateVersion: 0,
			Attributes: map[string]string{
				// EBS
				names.AttrRoleARN:     "arn:aws:iam::somenumber:role/tf_acctest_4271506651559170635", //lintignore:AWSAT005
				"s3_bucket_arn":       "arn:aws:s3:::tf-test-bucket",                                 //lintignore:AWSAT005
				"s3_buffer_interval":  "400",
				"s3_buffer_size":      acctest.Ct10,
				"s3_data_compression": "GZIP",
			},
			Expected: map[string]string{
				"s3_configuration.#":                    acctest.Ct1,
				"s3_configuration.0.bucket_arn":         "arn:aws:s3:::tf-test-bucket", //lintignore:AWSAT005
				"s3_configuration.0.buffer_interval":    "400",
				"s3_configuration.0.buffer_size":        acctest.Ct10,
				"s3_configuration.0.compression_format": "GZIP",
				"s3_configuration.0.role_arn":           "arn:aws:iam::somenumber:role/tf_acctest_4271506651559170635", //lintignore:AWSAT005
			},
		},
		"v0.6.16 and earlier, sparse": {
			StateVersion: 0,
			Attributes: map[string]string{
				// EBS
				names.AttrRoleARN: "arn:aws:iam::somenumber:role/tf_acctest_4271506651559170635", //lintignore:AWSAT005
				"s3_bucket_arn":   "arn:aws:s3:::tf-test-bucket",                                 //lintignore:AWSAT005
			},
			Expected: map[string]string{
				"s3_configuration.#":            acctest.Ct1,
				"s3_configuration.0.bucket_arn": "arn:aws:s3:::tf-test-bucket",                                 //lintignore:AWSAT005
				"s3_configuration.0.role_arn":   "arn:aws:iam::somenumber:role/tf_acctest_4271506651559170635", //lintignore:AWSAT005
			},
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         "i-abc123",
			Attributes: tc.Attributes,
		}
		is, err := tffirehose.MigrateState(
			tc.StateVersion, is, tc.Meta)

		if err != nil {
			t.Fatalf("bad: %s, err: %#v", tn, err)
		}

		for k, v := range tc.Expected {
			if is.Attributes[k] != v {
				t.Fatalf(
					"bad: %s\n\n expected: %#v -> %#v\n got: %#v -> %#v\n in: %#v",
					tn, k, v, k, is.Attributes[k], is.Attributes)
			}
		}
	}
}
