// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cloudwatch_event_rule", name="Rule")
func dataSourceRule() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRuleRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"event_bus_name": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  DefaultEventBusName,
			},
			"event_pattern": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrScheduleExpression: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	name := d.Get(names.AttrName).(string)
	eventBusName := d.Get("event_bus_name").(string)

	output, err := findRuleByTwoPartKey(ctx, conn, eventBusName, name)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge Rule (%s): %s", name, err)
	}

	arn := aws.ToString(output.Arn)
	d.SetId(name)
	d.Set(names.AttrARN, arn)
	d.Set("event_bus_name", output.EventBusName)
	if output.EventPattern != nil {
		pattern, err := ruleEventPatternJSONDecoder(aws.ToString(output.EventPattern))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		d.Set("event_pattern", pattern)
	}

	d.Set(names.AttrScheduleExpression, output.ScheduleExpression)
	d.Set(names.AttrState, output.State)

	return diags
}
