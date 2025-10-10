// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package pinpoint

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pinpoint"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pinpoint/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_pinpoint_sms_channel", name="SMS Channel")
func resourceSMSChannel() *schema.Resource {
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

func resourceSMSChannelUpsert(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	applicationId := d.Get(names.AttrApplicationID).(string)

	params := &awstypes.SMSChannelRequest{
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

	_, err := conn.UpdateSmsChannel(ctx, &req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting Pinpoint SMS Channel for application %s: %s", applicationId, err)
	}

	d.SetId(applicationId)

	return append(diags, resourceSMSChannelRead(ctx, d, meta)...)
}

func resourceSMSChannelRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	log.Printf("[INFO] Reading Pinpoint SMS Channel  for application %s", d.Id())

	output, err := findSMSChannelByApplicationId(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Pinpoint SMS Channel (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Pinpoint SMS Channel (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrApplicationID, output.ApplicationId)
	d.Set(names.AttrEnabled, output.Enabled)
	d.Set("sender_id", output.SenderId)
	d.Set("short_code", output.ShortCode)
	d.Set("promotional_messages_per_second", output.PromotionalMessagesPerSecond)
	d.Set("transactional_messages_per_second", output.TransactionalMessagesPerSecond)
	return diags
}

func resourceSMSChannelDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	log.Printf("[DEBUG] Deleting Pinpoint SMS Channel for application %s", d.Id())
	_, err := conn.DeleteSmsChannel(ctx, &pinpoint.DeleteSmsChannelInput{
		ApplicationId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Pinpoint SMS Channel for application %s: %s", d.Id(), err)
	}
	return diags
}

func findSMSChannelByApplicationId(ctx context.Context, conn *pinpoint.Client, applicationId string) (*awstypes.SMSChannelResponse, error) {
	input := &pinpoint.GetSmsChannelInput{
		ApplicationId: aws.String(applicationId),
	}

	output, err := conn.GetSmsChannel(ctx, input)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || output.SMSChannelResponse == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.SMSChannelResponse, nil
}
