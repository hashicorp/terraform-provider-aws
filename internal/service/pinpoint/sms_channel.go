// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpoint

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_pinpoint_sms_channel")
func ResourceSMSChannel() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSMSChannelUpsert,
		ReadWithoutTimeout:   resourceSMSChannelRead,
		UpdateWithoutTimeout: resourceSMSChannelUpsert,
		DeleteWithoutTimeout: resourceSMSChannelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrApplicationID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"sender_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"short_code": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"promotional_messages_per_second": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"transactional_messages_per_second": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceSMSChannelUpsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn(ctx)

	applicationId := d.Get(names.AttrApplicationID).(string)

	params := &pinpoint.SMSChannelRequest{
		Enabled: aws.Bool(d.Get(names.AttrEnabled).(bool)),
	}

	if v, ok := d.GetOk("sender_id"); ok {
		params.SenderId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("short_code"); ok {
		params.ShortCode = aws.String(v.(string))
	}

	req := pinpoint.UpdateSmsChannelInput{
		ApplicationId:     aws.String(applicationId),
		SMSChannelRequest: params,
	}

	_, err := conn.UpdateSmsChannelWithContext(ctx, &req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting Pinpoint SMS Channel for application %s: %s", applicationId, err)
	}

	d.SetId(applicationId)

	return append(diags, resourceSMSChannelRead(ctx, d, meta)...)
}

func resourceSMSChannelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn(ctx)

	log.Printf("[INFO] Reading Pinpoint SMS Channel  for application %s", d.Id())

	output, err := conn.GetSmsChannelWithContext(ctx, &pinpoint.GetSmsChannelInput{
		ApplicationId: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
			log.Printf("[WARN] Pinpoint SMS Channel for application %s not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "getting Pinpoint SMS Channel for application %s: %s", d.Id(), err)
	}

	res := output.SMSChannelResponse
	d.Set(names.AttrApplicationID, res.ApplicationId)
	d.Set(names.AttrEnabled, res.Enabled)
	d.Set("sender_id", res.SenderId)
	d.Set("short_code", res.ShortCode)
	d.Set("promotional_messages_per_second", res.PromotionalMessagesPerSecond)
	d.Set("transactional_messages_per_second", res.TransactionalMessagesPerSecond)
	return diags
}

func resourceSMSChannelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn(ctx)

	log.Printf("[DEBUG] Deleting Pinpoint SMS Channel for application %s", d.Id())
	_, err := conn.DeleteSmsChannelWithContext(ctx, &pinpoint.DeleteSmsChannelInput{
		ApplicationId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Pinpoint SMS Channel for application %s: %s", d.Id(), err)
	}
	return diags
}
