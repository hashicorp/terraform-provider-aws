// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAssociationRuleMigrateState(t *testing.T) {
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
			ID:           "test_document_association-dev",
			Attributes: map[string]string{
				names.AttrAssociationID: "fb03b7e6-4a21-4012-965f-91a38cfeec72",
				names.AttrInstanceID:    "i-0381b34d460caf6ef",
				names.AttrName:          "test_document_association-dev",
			},
			Expected: "fb03b7e6-4a21-4012-965f-91a38cfeec72",
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         tc.ID,
			Attributes: tc.Attributes,
		}
		is, err := associationMigrateState(tc.StateVersion, is, tc.Meta)

		if err != nil {
			t.Fatalf("bad: %s, err: %#v", tn, err)
		}

		if is.ID != tc.Expected {
			t.Fatalf("bad ssm association id: %s\n\n expected: %s", is.ID, tc.Expected)
		}
	}
}
