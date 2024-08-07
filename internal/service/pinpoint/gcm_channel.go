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

// @SDKResource("aws_pinpoint_gcm_channel", name="GCM Channel")
func resourceGCMChannel() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGCMChannelUpsert,
		ReadWithoutTimeout:   resourceGCMChannelRead,
		UpdateWithoutTimeout: resourceGCMChannelUpsert,
		DeleteWithoutTimeout: resourceGCMChannelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrApplicationID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"api_key": {
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

func resourceGCMChannelUpsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	applicationId := d.Get(names.AttrApplicationID).(string)

	params := &awstypes.GCMChannelRequest{}

	params.ApiKey = aws.String(d.Get("api_key").(string))
	params.Enabled = aws.Bool(d.Get(names.AttrEnabled).(bool))

	req := pinpoint.UpdateGcmChannelInput{
		ApplicationId:     aws.String(applicationId),
		GCMChannelRequest: params,
	}

	_, err := conn.UpdateGcmChannel(ctx, &req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting Pinpoint GCM Channel for application %s: %s", applicationId, err)
	}

	d.SetId(applicationId)

	return append(diags, resourceGCMChannelRead(ctx, d, meta)...)
}

func resourceGCMChannelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	log.Printf("[INFO] Reading Pinpoint GCM Channel for application %s", d.Id())

	output, err := findGCMChannelByApplicationId(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Pinpoint GCM Channel (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Pinpoint GCM Channel (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrApplicationID, output.ApplicationId)
	d.Set(names.AttrEnabled, output.Enabled)
	// api_key is never returned

	return diags
}

func resourceGCMChannelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	log.Printf("[DEBUG] Deleting Pinpoint GCM Channel for application %s", d.Id())
	_, err := conn.DeleteGcmChannel(ctx, &pinpoint.DeleteGcmChannelInput{
		ApplicationId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Pinpoint GCM Channel for application %s: %s", d.Id(), err)
	}
	return diags
}

func findGCMChannelByApplicationId(ctx context.Context, conn *pinpoint.Client, applicationId string) (*awstypes.GCMChannelResponse, error) {
	input := &pinpoint.GetGcmChannelInput{
		ApplicationId: aws.String(applicationId),
	}

	output, err := conn.GetGcmChannel(ctx, input)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || output.GCMChannelResponse == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.GCMChannelResponse, nil
}
