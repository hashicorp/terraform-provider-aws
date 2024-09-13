// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesis

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func resourceStreamResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"shard_count": {
				Type:     schema.TypeInt,
				Required: true,
			},

			names.AttrRetentionPeriod: {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  24,
			},

			"shard_level_metrics": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"encryption_type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "NONE",
			},

			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Optional: true,
			},

			names.AttrARN: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchema(),
		},
	}
}

func StreamStateUpgradeV0(_ context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	rawState["enforce_consumer_deletion"] = false

	return rawState, nil
}
