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
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_pinpoint_email_channel", name="Email Channel")
func resourceEmailChannel() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEmailChannelUpsert,
		ReadWithoutTimeout:   resourceEmailChannelRead,
		UpdateWithoutTimeout: resourceEmailChannelUpsert,
		DeleteWithoutTimeout: resourceEmailChannelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrApplicationID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"configuration_set": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"from_address": {
				Type:     schema.TypeString,
				Required: true,
			},
			"identity": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"orchestration_sending_role_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"messages_per_second": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceEmailChannelUpsert(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	applicationId := d.Get(names.AttrApplicationID).(string)

	params := &awstypes.EmailChannelRequest{}

	params.Enabled = aws.Bool(d.Get(names.AttrEnabled).(bool))
	params.FromAddress = aws.String(d.Get("from_address").(string))
	params.Identity = aws.String(d.Get("identity").(string))

	if v, ok := d.GetOk("orchestration_sending_role_arn"); ok {
		params.OrchestrationSendingRoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrRoleARN); ok {
		params.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("configuration_set"); ok {
		params.ConfigurationSet = aws.String(v.(string))
	}

	req := pinpoint.UpdateEmailChannelInput{
		ApplicationId:       aws.String(applicationId),
		EmailChannelRequest: params,
	}

	_, err := conn.UpdateEmailChannel(ctx, &req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Pinpoint Email Channel for application %s: %s", applicationId, err)
	}

	d.SetId(applicationId)

	return append(diags, resourceEmailChannelRead(ctx, d, meta)...)
}

func resourceEmailChannelRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	log.Printf("[INFO] Reading Pinpoint Email Channel for application %s", d.Id())

	output, err := findEmailChannelByApplicationId(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Pinpoint Email Channel (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Pinpoint Email Channel (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrApplicationID, output.ApplicationId)
	d.Set(names.AttrEnabled, output.Enabled)
	d.Set("from_address", output.FromAddress)
	d.Set("identity", output.Identity)
	d.Set("orchestration_sending_role_arn", output.OrchestrationSendingRoleArn)
	d.Set(names.AttrRoleARN, output.RoleArn)
	d.Set("configuration_set", output.ConfigurationSet)
	d.Set("messages_per_second", output.MessagesPerSecond)

	return diags
}

func resourceEmailChannelDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	log.Printf("[DEBUG] Deleting Pinpoint Email Channel for application %s", d.Id())
	_, err := conn.DeleteEmailChannel(ctx, &pinpoint.DeleteEmailChannelInput{
		ApplicationId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Pinpoint Email Channel for application %s: %s", d.Id(), err)
	}
	return diags
}

func findEmailChannelByApplicationId(ctx context.Context, conn *pinpoint.Client, applicationId string) (*awstypes.EmailChannelResponse, error) {
	input := &pinpoint.GetEmailChannelInput{
		ApplicationId: aws.String(applicationId),
	}

	output, err := conn.GetEmailChannel(ctx, input)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || output.EmailChannelResponse == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.EmailChannelResponse, nil
}
