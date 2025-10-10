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

// @SDKResource("aws_pinpoint_adm_channel", name="ADM Channel")
func resourceADMChannel() *schema.Resource {
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

func resourceADMChannelUpsert(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	applicationId := d.Get(names.AttrApplicationID).(string)

	params := &awstypes.ADMChannelRequest{}

	params.ClientId = aws.String(d.Get(names.AttrClientID).(string))
	params.ClientSecret = aws.String(d.Get(names.AttrClientSecret).(string))
	params.Enabled = aws.Bool(d.Get(names.AttrEnabled).(bool))

	req := pinpoint.UpdateAdmChannelInput{
		ApplicationId:     aws.String(applicationId),
		ADMChannelRequest: params,
	}

	_, err := conn.UpdateAdmChannel(ctx, &req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Pinpoint ADM Channel: %s", err)
	}

	d.SetId(applicationId)

	return append(diags, resourceADMChannelRead(ctx, d, meta)...)
}

func resourceADMChannelRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	log.Printf("[INFO] Reading Pinpoint ADM Channel for application %s", d.Id())

	output, err := findADMChannelByApplicationId(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Pinpoint ADM Channel (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Pinpoint ADM Channel (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrApplicationID, output.ApplicationId)
	d.Set(names.AttrEnabled, output.Enabled)
	// client_id and client_secret are never returned

	return diags
}

func resourceADMChannelDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	log.Printf("[DEBUG] Pinpoint Delete ADM Channel: %s", d.Id())
	_, err := conn.DeleteAdmChannel(ctx, &pinpoint.DeleteAdmChannelInput{
		ApplicationId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Pinpoint ADM Channel for application %s: %s", d.Id(), err)
	}
	return diags
}

func findADMChannelByApplicationId(ctx context.Context, conn *pinpoint.Client, applicationId string) (*awstypes.ADMChannelResponse, error) {
	input := &pinpoint.GetAdmChannelInput{
		ApplicationId: aws.String(applicationId),
	}

	output, err := conn.GetAdmChannel(ctx, input)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || output.ADMChannelResponse == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ADMChannelResponse, nil
}
