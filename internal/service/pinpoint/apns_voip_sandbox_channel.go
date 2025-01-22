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

// @SDKResource("aws_pinpoint_apns_voip_sandbox_channel", name="APNS VoIP Sandbox Channel")
func resourceAPNSVoIPSandboxChannel() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAPNSVoIPSandboxChannelUpsert,
		ReadWithoutTimeout:   resourceAPNSVoIPSandboxChannelRead,
		UpdateWithoutTimeout: resourceAPNSVoIPSandboxChannelUpsert,
		DeleteWithoutTimeout: resourceAPNSVoIPSandboxChannelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrApplicationID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"bundle_id": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			names.AttrCertificate: {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"default_authentication_method": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			names.AttrPrivateKey: {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"team_id": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"token_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"token_key_id": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceAPNSVoIPSandboxChannelUpsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	certificate, certificateOk := d.GetOk(names.AttrCertificate)
	privateKey, privateKeyOk := d.GetOk(names.AttrPrivateKey)

	bundleId, bundleIdOk := d.GetOk("bundle_id")
	teamId, teamIdOk := d.GetOk("team_id")
	tokenKey, tokenKeyOk := d.GetOk("token_key")
	tokenKeyId, tokenKeyIdOk := d.GetOk("token_key_id")

	if !(certificateOk && privateKeyOk) && !(bundleIdOk && teamIdOk && tokenKeyOk && tokenKeyIdOk) {
		return sdkdiag.AppendErrorf(diags, "At least one set of credentials is required; either [certificate, private_key] or [bundle_id, team_id, token_key, token_key_id]")
	}

	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	applicationId := d.Get(names.AttrApplicationID).(string)

	params := &awstypes.APNSVoipSandboxChannelRequest{}

	params.DefaultAuthenticationMethod = aws.String(d.Get("default_authentication_method").(string))
	params.Enabled = aws.Bool(d.Get(names.AttrEnabled).(bool))

	params.Certificate = aws.String(certificate.(string))
	params.PrivateKey = aws.String(privateKey.(string))

	params.BundleId = aws.String(bundleId.(string))
	params.TeamId = aws.String(teamId.(string))
	params.TokenKey = aws.String(tokenKey.(string))
	params.TokenKeyId = aws.String(tokenKeyId.(string))

	req := pinpoint.UpdateApnsVoipSandboxChannelInput{
		ApplicationId:                 aws.String(applicationId),
		APNSVoipSandboxChannelRequest: params,
	}

	_, err := conn.UpdateApnsVoipSandboxChannel(ctx, &req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Pinpoint APNS VoIP Sandbox Channel for Application %s: %s", applicationId, err)
	}

	d.SetId(applicationId)

	return append(diags, resourceAPNSVoIPSandboxChannelRead(ctx, d, meta)...)
}

func resourceAPNSVoIPSandboxChannelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	log.Printf("[INFO] Reading Pinpoint APNS VoIP Sandbox Channel for Application %s", d.Id())

	output, err := findAPNSVoIPSandboxChannelByApplicationId(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Pinpoint APNS VoIP Sandbox Channel (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Pinpoint APNS VoIP Sandbox Channel (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrApplicationID, output.ApplicationId)
	d.Set("default_authentication_method", output.DefaultAuthenticationMethod)
	d.Set(names.AttrEnabled, output.Enabled)
	// Sensitive params are not returned

	return diags
}

func resourceAPNSVoIPSandboxChannelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointClient(ctx)

	log.Printf("[DEBUG] Deleting Pinpoint APNS VoIP Sandbox Channel: %s", d.Id())
	_, err := conn.DeleteApnsVoipSandboxChannel(ctx, &pinpoint.DeleteApnsVoipSandboxChannelInput{
		ApplicationId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Pinpoint APNS VoIP Sandbox Channel for Application %s: %s", d.Id(), err)
	}
	return diags
}

func findAPNSVoIPSandboxChannelByApplicationId(ctx context.Context, conn *pinpoint.Client, applicationId string) (*awstypes.APNSVoipSandboxChannelResponse, error) {
	input := &pinpoint.GetApnsVoipSandboxChannelInput{
		ApplicationId: aws.String(applicationId),
	}

	output, err := conn.GetApnsVoipSandboxChannel(ctx, input)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if output == nil || output.APNSVoipSandboxChannelResponse == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.APNSVoipSandboxChannelResponse, nil
}
