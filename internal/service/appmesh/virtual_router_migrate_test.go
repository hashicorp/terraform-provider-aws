// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appmesh_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfappmesh "github.com/hashicorp/terraform-provider-aws/internal/service/appmesh"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestVirtualRouterMigrateState(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
		Meta         interface{}
	}{
		"v0_1-emptySpec": {
			StateVersion: 0,
			Attributes: map[string]string{
				names.AttrName: "svcb",
				"spec.#":       acctest.Ct1,
			},
			Expected: map[string]string{
				names.AttrName: "svcb",
				"spec.#":       acctest.Ct1,
			},
		},
		"v0_1-nonEmptySpec": {
			StateVersion: 0,
			Attributes: map[string]string{
				names.AttrName:                   "svcb",
				"spec.#":                         acctest.Ct1,
				"spec.0.service_names.#":         acctest.Ct1,
				"spec.0.service_names.423761483": "serviceb.simpleapp.local",
			},
			Expected: map[string]string{
				names.AttrName: "svcb",
				"spec.#":       acctest.Ct1,
			},
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         "some_id",
			Attributes: tc.Attributes,
		}

		is, err := tfappmesh.ResourceVirtualRouter().MigrateState(tc.StateVersion, is, tc.Meta)
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
