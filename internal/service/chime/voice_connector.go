// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chime

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_chime_voice_connector", name="Voice Connector")
// @Tags(identifierAttribute="arn")
func ResourceVoiceConnector() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVoiceConnectorCreate,
		ReadWithoutTimeout:   resourceVoiceConnectorRead,
		UpdateWithoutTimeout: resourceVoiceConnectorUpdate,
		DeleteWithoutTimeout: resourceVoiceConnectorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_region": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				Default:      chime.VoiceConnectorAwsRegionUsEast1,
				ValidateFunc: validation.StringInSlice([]string{chime.VoiceConnectorAwsRegionUsEast1, chime.VoiceConnectorAwsRegionUsWest2}, false),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"outbound_host_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"require_encryption": {
				Type:     schema.TypeBool,
				Required: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVoiceConnectorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeConn(ctx)

	createInput := &chime.CreateVoiceConnectorInput{
		Name:              aws.String(d.Get("name").(string)),
		RequireEncryption: aws.Bool(d.Get("require_encryption").(bool)),
	}

	if v, ok := d.GetOk("aws_region"); ok {
		createInput.AwsRegion = aws.String(v.(string))
	}

	resp, err := conn.CreateVoiceConnectorWithContext(ctx, createInput)
	if err != nil || resp.VoiceConnector == nil {
		return sdkdiag.AppendErrorf(diags, "creating Chime Voice connector: %s", err)
	}

	d.SetId(aws.StringValue(resp.VoiceConnector.VoiceConnectorId))

	tagsConn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)
	if err := createTags(ctx, tagsConn, aws.StringValue(resp.VoiceConnector.VoiceConnectorArn), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Chime Voice Connector (%s) tags: %s", d.Id(), err)
	}

	return append(diags, resourceVoiceConnectorRead(ctx, d, meta)...)
}

func resourceVoiceConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeConn(ctx)

	getInput := &chime.GetVoiceConnectorInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	resp, err := conn.GetVoiceConnectorWithContext(ctx, getInput)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, chime.ErrCodeNotFoundException) {
		log.Printf("[WARN] Chime Voice connector %s not found", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil || resp.VoiceConnector == nil {
		return sdkdiag.AppendErrorf(diags, "getting Voice connector (%s): %s", d.Id(), err)
	}

	d.Set("arn", resp.VoiceConnector.VoiceConnectorArn)
	d.Set("aws_region", resp.VoiceConnector.AwsRegion)
	d.Set("outbound_host_name", resp.VoiceConnector.OutboundHostName)
	d.Set("require_encryption", resp.VoiceConnector.RequireEncryption)
	d.Set("name", resp.VoiceConnector.Name)

	return diags
}

func resourceVoiceConnectorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeConn(ctx)

	if d.HasChanges("name", "require_encryption") {
		updateInput := &chime.UpdateVoiceConnectorInput{
			VoiceConnectorId:  aws.String(d.Id()),
			Name:              aws.String(d.Get("name").(string)),
			RequireEncryption: aws.Bool(d.Get("require_encryption").(bool)),
		}

		if _, err := conn.UpdateVoiceConnectorWithContext(ctx, updateInput); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Voice connector (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceVoiceConnectorRead(ctx, d, meta)...)
}

func resourceVoiceConnectorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeConn(ctx)

	input := &chime.DeleteVoiceConnectorInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	if _, err := conn.DeleteVoiceConnectorWithContext(ctx, input); err != nil {
		if tfawserr.ErrCodeEquals(err, chime.ErrCodeNotFoundException) {
			log.Printf("[WARN] Chime Voice connector %s not found", d.Id())
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Voice connector (%s)", d.Id())
	}

	return diags
}
