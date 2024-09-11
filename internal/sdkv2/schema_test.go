// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceElemFromResourceElem(t *testing.T) {
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

	got := DataSourceElemFromResourceElem(input)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}
