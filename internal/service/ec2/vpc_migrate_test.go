// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestVPCMigrateState(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		StateVersion int
		ID           string
		Attributes   map[string]string
		Expected     string
		Meta         interface{}
	}{
		"v0_1": {
			StateVersion: 0,
			ID:           "some_id",
			Attributes: map[string]string{
				"assign_generated_ipv6_cidr_block": acctest.CtTrue,
			},
			Expected: acctest.CtFalse,
		},
		"v0_1_without_value": {
			StateVersion: 0,
			ID:           "some_id",
			Attributes:   map[string]string{},
			Expected:     acctest.CtFalse,
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         tc.ID,
			Attributes: tc.Attributes,
		}
		is, err := tfec2.VPCMigrateState(
			tc.StateVersion, is, tc.Meta)

		if err != nil {
			t.Fatalf("bad: %s, err: %#v", tn, err)
		}

		if is.Attributes["assign_generated_ipv6_cidr_block"] != tc.Expected {
			t.Fatalf("bad VPC Migrate: %s\n\n expected: %s", is.Attributes["assign_generated_ipv6_cidr_block"], tc.Expected)
		}
	}
}
