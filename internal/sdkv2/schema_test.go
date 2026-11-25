// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestComputedOnlyFromResource(t *testing.T) {
	t.Parallel()

	input := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key1": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key2": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key1": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"key2": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeInt},
						},
					},
				},
			},
		},
	}
	want := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key1": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key2": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key1": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"key2": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeInt},
						},
					},
				},
			},
		},
	}

	got := ComputedOnlyFromResource(input)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestWarnShouldBeComputedOnly(t *testing.T) {
	t.Parallel()

	path := cty.GetAttrPath("arn")
	diags := WarnShouldBeComputedOnly("x", path)
	if len(diags) != 1 || diags[0].Severity != diag.Warning || !diags[0].AttributePath.Equals(path) {
		t.Errorf("got %v, want one warning at %v", diags, path)
	}
}

func TestWarnShouldBeComputedOnlyRawConfig(t *testing.T) {
	t.Parallel()

	ty := cty.Object(map[string]cty.Type{"members": cty.Set(cty.String), "name": cty.String})
	testCases := map[string]struct {
		config       cty.Value
		wantWarnings int
	}{
		"set": {
			config:       cty.ObjectVal(map[string]cty.Value{"members": cty.SetVal([]cty.Value{cty.StringVal("a")}), "name": cty.NullVal(cty.String)}),
			wantWarnings: 1,
		},
		"empty set": {
			config:       cty.ObjectVal(map[string]cty.Value{"members": cty.SetValEmpty(cty.String), "name": cty.NullVal(cty.String)}),
			wantWarnings: 1,
		},
		"unknown": {
			config:       cty.ObjectVal(map[string]cty.Value{"members": cty.UnknownVal(cty.Set(cty.String)), "name": cty.NullVal(cty.String)}),
			wantWarnings: 1,
		},
		"not set": {
			config: cty.ObjectVal(map[string]cty.Value{"members": cty.NullVal(cty.Set(cty.String)), "name": cty.StringVal("n")}),
		},
		"null config": {
			config: cty.NullVal(ty),
		},
		"unknown config": {
			config: cty.UnknownVal(ty),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var resp schema.ValidateResourceConfigFuncResponse
			WarnShouldBeComputedOnlyRawConfig("members")(t.Context(), schema.ValidateResourceConfigFuncRequest{RawConfig: tc.config}, &resp)
			if got := len(resp.Diagnostics); got != tc.wantWarnings {
				t.Fatalf("got %d diagnostics, want %d", got, tc.wantWarnings)
			}
			for _, d := range resp.Diagnostics {
				if d.Severity != diag.Warning || !d.AttributePath.Equals(cty.GetAttrPath("members")) {
					t.Errorf("got %v, want a warning at members", d)
				}
			}
		})
	}
}
