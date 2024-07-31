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

// @SDKResource("aws_pinpoint_apns_sandbox_channel")
func ResourceAPNSSandboxChannel() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAPNSSandboxChannelUpsert,
		ReadWithoutTimeout:   resourceAPNSSandboxChannelRead,
		UpdateWithoutTimeout: resourceAPNSSandboxChannelUpsert,
		DeleteWithoutTimeout: resourceAPNSSandboxChannelDelete,
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

func resourceAPNSSandboxChannelUpsert(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	conn := meta.(*conns.AWSClient).PinpointConn(ctx)

	applicationId := d.Get(names.AttrApplicationID).(string)

	params := &pinpoint.APNSSandboxChannelRequest{}

	params.DefaultAuthenticationMethod = aws.String(d.Get("default_authentication_method").(string))
	params.Enabled = aws.Bool(d.Get(names.AttrEnabled).(bool))

	params.Certificate = aws.String(certificate.(string))
	params.PrivateKey = aws.String(privateKey.(string))

	params.BundleId = aws.String(bundleId.(string))
	params.TeamId = aws.String(teamId.(string))
	params.TokenKey = aws.String(tokenKey.(string))
	params.TokenKeyId = aws.String(tokenKeyId.(string))

	req := pinpoint.UpdateApnsSandboxChannelInput{
		ApplicationId:             aws.String(applicationId),
		APNSSandboxChannelRequest: params,
	}

	_, err := conn.UpdateApnsSandboxChannelWithContext(ctx, &req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Pinpoint APNs Sandbox Channel for Application %s: %s", applicationId, err)
	}

	d.SetId(applicationId)

	return append(diags, resourceAPNSSandboxChannelRead(ctx, d, meta)...)
}

func resourceAPNSSandboxChannelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn(ctx)

	log.Printf("[INFO] Reading Pinpoint APNs Channel for Application %s", d.Id())

	output, err := conn.GetApnsSandboxChannelWithContext(ctx, &pinpoint.GetApnsSandboxChannelInput{
		ApplicationId: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
			log.Printf("[WARN] Pinpoint APNs Sandbox Channel for application %s not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "getting Pinpoint APNs Sandbox Channel for application %s: %s", d.Id(), err)
	}

	d.Set(names.AttrApplicationID, output.APNSSandboxChannelResponse.ApplicationId)
	d.Set("default_authentication_method", output.APNSSandboxChannelResponse.DefaultAuthenticationMethod)
	d.Set(names.AttrEnabled, output.APNSSandboxChannelResponse.Enabled)
	// Sensitive params are not returned

	return diags
}

func resourceAPNSSandboxChannelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).PinpointConn(ctx)

	log.Printf("[DEBUG] Deleting Pinpoint APNs Sandbox Channel: %s", d.Id())
	_, err := conn.DeleteApnsSandboxChannelWithContext(ctx, &pinpoint.DeleteApnsSandboxChannelInput{
		ApplicationId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, pinpoint.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Pinpoint APNs Sandbox Channel for Application %s: %s", d.Id(), err)
	}
	return diags
}
