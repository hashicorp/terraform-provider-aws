// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package medialive

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func destinationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"destination_ref_id": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func connectionRetryIntervalSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
	}
}

func filecacheDurationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
	}
}

func numRetriesSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
	}
}

func restartDelaySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeInt,
		Optional: true,
	}
}

func inputLocationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrURI: {
					Type:     schema.TypeString,
					Required: true,
				},
				"password_param": {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
				names.AttrUsername: {
					Type:     schema.TypeString,
					Optional: true,
					Computed: true,
				},
			},
		},
	}
}
