// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcegroups

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func resourceResourceConfigV0() *schema.Resource {
	// Resource with v0 schema (provider v5.81.0 and below)
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"group_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrResourceARN: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrResourceType: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceStateUpgradeV0(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	if rawState == nil {
		rawState = map[string]any{}
	}

	// Convert id to comma-delimited string combining group_arn and resource_arn
	parts := []string{
		rawState["group_arn"].(string),
		rawState[names.AttrResourceARN].(string),
	}

	id, err := flex.FlattenResourceId(parts, resourceIDPartCount, false)
	if err != nil {
		return rawState, err
	}
	rawState[names.AttrID] = id

	return rawState, nil
}
