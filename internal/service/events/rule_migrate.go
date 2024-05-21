// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func resourceRuleV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"event_bus_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  DefaultEventBusName,
			},
			"event_pattern": {
				Type:     schema.TypeString,
				Optional: true,
				StateFunc: func(v interface{}) string {
					json, _ := ruleEventPatternJSONDecoder(v.(string))
					return json
				},
			},
			"is_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrNamePrefix: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrScheduleExpression: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceRuleUpgradeV0(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	if rawState == nil {
		rawState = map[string]any{}
	}

	tflog.Debug(ctx, "Upgrading resource", map[string]any{
		"from_version": 0,
		"to_version":   1,
	})

	if rawState["is_enabled"].(bool) {
		rawState[names.AttrState] = types.RuleStateEnabled
	} else {
		rawState[names.AttrState] = types.RuleStateDisabled
	}

	return rawState, nil
}
