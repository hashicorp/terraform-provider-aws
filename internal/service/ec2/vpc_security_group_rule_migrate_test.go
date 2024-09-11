// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestSecurityGroupRuleMigrateState(t *testing.T) {
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
			ID:           "sg-4235098228",
			Attributes: map[string]string{
				"self":                     acctest.CtFalse,
				"to_port":                  acctest.Ct0,
				"security_group_id":        "sg-13877277",
				"cidr_blocks.#":            acctest.Ct0,
				names.AttrType:             "ingress",
				names.AttrProtocol:         "-1",
				"from_port":                acctest.Ct0,
				"source_security_group_id": "sg-11877275",
			},
			Expected: "sgrule-2889201120",
		},
		"v0_2": {
			StateVersion: 0,
			ID:           "sg-1021609891",
			Attributes: map[string]string{
				"security_group_id": "sg-0981746d",
				"from_port":         acctest.Ct0,
				"to_port":           acctest.Ct0,
				names.AttrType:      "ingress",
				"self":              acctest.CtFalse,
				names.AttrProtocol:  "-1",
				"cidr_blocks.0":     "172.16.1.0/24",
				"cidr_blocks.1":     "172.16.2.0/24",
				"cidr_blocks.2":     "172.16.3.0/24",
				"cidr_blocks.3":     "172.16.4.0/24",
				"cidr_blocks.#":     acctest.Ct4},
			Expected: "sgrule-1826358977",
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         tc.ID,
			Attributes: tc.Attributes,
		}
		is, err := tfec2.SecurityGroupRuleMigrateState(
			tc.StateVersion, is, tc.Meta)

		if err != nil {
			t.Fatalf("bad: %s, err: %#v", tn, err)
		}

		if is.ID != tc.Expected {
			t.Fatalf("bad sg rule id: %s\n\n expected: %s", is.ID, tc.Expected)
		}
	}
}
