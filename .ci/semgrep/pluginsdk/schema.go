// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceResource() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCCreate,
		ReadWithoutTimeout:   resourceVPCRead,
		UpdateWithoutTimeout: resourceVPCUpdate,
		DeleteWithoutTimeout: resourceVPCDelete,

		// ruleid: use-schema-func
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceNested() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCCreate,
		ReadWithoutTimeout:   resourceVPCRead,
		UpdateWithoutTimeout: resourceVPCUpdate,
		DeleteWithoutTimeout: resourceVPCDelete,

		// ruleid: use-schema-func
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"list": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					// ok: use-schema-func
					Schema: map[string]*schema.Schema{
						"item": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}
