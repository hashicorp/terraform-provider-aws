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

// @SDKResource("aws_pinpoint_adm_channel")
func ResourceADMChannel() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceADMChannelUpsert,
		ReadWithoutTimeout:   resourceADMChannelRead,
		UpdateWithoutTimeout: resourceADMChannelUpsert,
		DeleteWithoutTimeout: resourceADMChannelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrApplicationID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrClientID: {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			names.AttrClientSecret: {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
		},
	}
}

func resourceADMChannelUpsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn(ctx)

	applicationId := d.Get(names.AttrApplicationID).(string)

	params := &pinpoint.ADMChannelRequest{}

	params.ClientId = aws.String(d.Get(names.AttrClientID).(string))
	params.ClientSecret = aws.String(d.Get(names.AttrClientSecret).(string))
	params.Enabled = aws.Bool(d.Get(names.AttrEnabled).(bool))

	req := pinpoint.UpdateAdmChannelInput{
		ApplicationId:     aws.String(applicationId),
		ADMChannelRequest: params,
	}

	_, err := conn.UpdateAdmChannelWithContext(ctx, &req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Pinpoint ADM Channel: %s", err)
	}

	d.SetId(applicationId)

	return append(diags, resourceADMChannelRead(ctx, d, meta)...)
}

func resourceADMChannelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn(ctx)

	log.Printf("[INFO] Reading Pinpoint ADM Channel for application %s", d.Id())

	channel, err := conn.GetAdmChannelWithContext(ctx, &pinpoint.GetAdmChannelInput{
		ApplicationId: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
			log.Printf("[WARN] Pinpoint ADM Channel for application %s not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "getting Pinpoint ADM Channel for application %s: %s", d.Id(), err)
	}

	d.Set(names.AttrApplicationID, channel.ADMChannelResponse.ApplicationId)
	d.Set(names.AttrEnabled, channel.ADMChannelResponse.Enabled)
	// client_id and client_secret are never returned

	return diags
}

func resourceADMChannelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn(ctx)

	log.Printf("[DEBUG] Pinpoint Delete ADM Channel: %s", d.Id())
	_, err := conn.DeleteAdmChannelWithContext(ctx, &pinpoint.DeleteAdmChannelInput{
		ApplicationId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Pinpoint ADM Channel for application %s: %s", d.Id(), err)
	}
	return diags
}
