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

// @SDKResource("aws_pinpoint_baidu_channel", name="Baidu Channel")
func resourceBaiduChannel() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBaiduChannelUpsert,
		ReadWithoutTimeout:   resourceBaiduChannelRead,
		UpdateWithoutTimeout: resourceBaiduChannelUpsert,
		DeleteWithoutTimeout: resourceBaiduChannelDelete,
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
			"api_key": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			names.AttrSecretKey: {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceBaiduChannelUpsert(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	applicationId := d.Get(names.AttrApplicationID).(string)

	params := &awstypes.BaiduChannelRequest{}

	params.Enabled = aws.Bool(d.Get(names.AttrEnabled).(bool))
	params.ApiKey = aws.String(d.Get("api_key").(string))
	params.SecretKey = aws.String(d.Get(names.AttrSecretKey).(string))

	req := pinpoint.UpdateBaiduChannelInput{
		ApplicationId:       aws.String(applicationId),
		BaiduChannelRequest: params,
	}

	_, err := conn.UpdateBaiduChannel(ctx, &req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Pinpoint Baidu Channel for application %s: %s", applicationId, err)
	}

	d.SetId(applicationId)

	return append(diags, resourceBaiduChannelRead(ctx, d, meta)...)
}

func resourceBaiduChannelRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	log.Printf("[INFO] Reading Pinpoint Baidu Channel for application %s", d.Id())

	output, err := findBaiduChannelByApplicationId(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Pinpoint Baidu Channel (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Pinpoint Baidu Channel (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrApplicationID, output.ApplicationId)
	d.Set(names.AttrEnabled, output.Enabled)
	// ApiKey and SecretKey are never returned

	return diags
}

func resourceBaiduChannelDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	log.Printf("[DEBUG] Deleting Pinpoint Baidu Channel for application %s", d.Id())
	_, err := conn.DeleteBaiduChannel(ctx, &pinpoint.DeleteBaiduChannelInput{
		ApplicationId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Pinpoint Baidu Channel for application %s: %s", d.Id(), err)
	}
	return diags
}

func findBaiduChannelByApplicationId(ctx context.Context, conn *pinpoint.Client, applicationId string) (*awstypes.BaiduChannelResponse, error) {
	input := &pinpoint.GetBaiduChannelInput{
		ApplicationId: aws.String(applicationId),
	}

	output, err := conn.GetBaiduChannel(ctx, input)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || output.BaiduChannelResponse == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.BaiduChannelResponse, nil
}
