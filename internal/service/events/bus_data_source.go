// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_cloudwatch_event_bus", name="Event Bus")
func dataSourceBus() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceBusRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dead_letter_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_key_identifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"log_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"include_detail": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"level": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceBusRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EventsClient(ctx)

	eventBusName := d.Get(names.AttrName).(string)
	output, err := findEventBusByName(ctx, conn, eventBusName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EventBridge Event Bus (%s): %s", eventBusName, err)
	}

	d.SetId(eventBusName)
	d.Set(names.AttrARN, output.Arn)
	if err := d.Set("dead_letter_config", flattenDeadLetterConfig(output.DeadLetterConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting dead_letter_config: %s", err)
	}
	d.Set(names.AttrDescription, output.Description)
	d.Set("kms_key_identifier", output.KmsKeyIdentifier)
	if err := d.Set("log_config", flattenLogConfig(output.LogConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting log_config: %s", err)
	}
	d.Set(names.AttrName, output.Name)

	return diags
}
