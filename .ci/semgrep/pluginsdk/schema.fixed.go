// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// IMPORTANT: The "fixed" file must not be formatted with gofmt.
// Semgrep does not handle formatting of multiline fixes in Go correctly.

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
		SchemaFunc: func() map[string]*schema.Schema {
  	return map[string]*schema.Schema{
  			"name": {
  				Type:     schema.TypeString,
  				Required: true,
  				ForceNew: true,
  			},
  		}
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
		SchemaFunc: func() map[string]*schema.Schema {
  	return map[string]*schema.Schema{
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
  		}
  },
	}
}

func resourcePreamble() *schema.Resource {
	x := somevalue()

	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCCreate,
		ReadWithoutTimeout:   resourceVPCRead,
		UpdateWithoutTimeout: resourceVPCUpdate,
		DeleteWithoutTimeout: resourceVPCDelete,

		// ruleid: use-schema-func
		SchemaFunc: func() map[string]*schema.Schema {
  	return map[string]*schema.Schema{
  			"name": {
  				Type:     schema.TypeString,
  				Required: true,
  				ForceNew: true,
  			},
  		}
  },
	}
}

func resourceMigratorV1() *schema.Resource {
	x := somevalue()

	return &schema.Resource{
		// ok: use-schema-func
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}
