// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chime

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chimesdkvoice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
				Computed:     true,
				ValidateFunc: validation.StringInSlice(chimesdkvoice.VoiceConnectorAwsRegion_Values(), false),
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

		CustomizeDiff: customdiff.All(
			verify.SetTagsDiff,
			resourceVoiceConnectorDefaultRegion,
		),
	}
}

func resourceVoiceConnectorDefaultRegion(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if v, ok := diff.Get("aws_region").(string); !ok || v == "" {
		if err := diff.SetNew("aws_region", meta.(*conns.AWSClient).Region); err != nil {
			return err
		}
	}

	return nil
}

func resourceVoiceConnectorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	createInput := &chimesdkvoice.CreateVoiceConnectorInput{
		Name:              aws.String(d.Get("name").(string)),
		RequireEncryption: aws.Bool(d.Get("require_encryption").(bool)),
		Tags:              getTagsIn(ctx),
	}

	if v, ok := d.GetOk("aws_region"); ok {
		createInput.AwsRegion = aws.String(v.(string))
	} else {
		createInput.AwsRegion = aws.String(meta.(*conns.AWSClient).Region)
	}

	resp, err := conn.CreateVoiceConnectorWithContext(ctx, createInput)
	if err != nil || resp.VoiceConnector == nil {
		return sdkdiag.AppendErrorf(diags, "creating Chime Voice connector: %s", err)
	}

	d.SetId(aws.StringValue(resp.VoiceConnector.VoiceConnectorId))

	//tagsConn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)
	//if err := createTags(ctx, tagsConn, aws.StringValue(resp.VoiceConnector.VoiceConnectorArn), getTagsIn(ctx)); err != nil {
	//	return sdkdiag.AppendErrorf(diags, "setting Chime Voice Connector (%s) tags: %s", d.Id(), err)
	//}

	return append(diags, resourceVoiceConnectorRead(ctx, d, meta)...)
}

func resourceVoiceConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	//var resp *chimesdkvoice.VoiceConnector
	//err := tfresource.Retry(ctx, 1*time.Minute, func() *retry.RetryError {
	//	var err error
	//	resp, err = findVoiceConnectorByID(ctx, conn, d.Id())
	//	if d.IsNewResource() && tfresource.NotFound(err) {
	//		return retry.RetryableError(err)
	//	}
	//
	//	if err != nil {
	//		return retry.NonRetryableError(err)
	//	}
	//
	//	return nil
	//}, tfresource.WithDelay(5*time.Second))

	resp, err := findVoiceConnectorWithRetry(ctx, conn, d.IsNewResource(), d.Id())

	if tfresource.TimedOut(err) {
		resp, err = findVoiceConnectorByID(ctx, conn, d.Id())
	}
	//outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, 1*time.Minute, func() (interface{}, error) {
	//	return findVoiceConnectorByID(ctx, conn, d.Id())
	//}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Chime Voice connector %s not found", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Voice Connector (%s): %s", d.Id(), err)
	}

	// resp := outputRaw.(*chimesdkvoice.VoiceConnector)

	d.Set("arn", resp.VoiceConnectorArn)
	d.Set("aws_region", resp.AwsRegion)
	d.Set("outbound_host_name", resp.OutboundHostName)
	d.Set("require_encryption", resp.RequireEncryption)
	d.Set("name", resp.Name)

	return diags
}

func resourceVoiceConnectorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	if d.HasChanges("name", "require_encryption") {
		updateInput := &chimesdkvoice.UpdateVoiceConnectorInput{
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
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceConn(ctx)

	input := &chimesdkvoice.DeleteVoiceConnectorInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	if _, err := conn.DeleteVoiceConnectorWithContext(ctx, input); err != nil {
		if tfawserr.ErrCodeEquals(err, chimesdkvoice.ErrCodeNotFoundException) {
			log.Printf("[WARN] Chime Voice connector %s not found", d.Id())
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Voice connector (%s)", d.Id())
	}

	return diags
}

func findVoiceConnectorByID(ctx context.Context, conn *chimesdkvoice.ChimeSDKVoice, id string) (*chimesdkvoice.VoiceConnector, error) {
	in := &chimesdkvoice.GetVoiceConnectorInput{
		VoiceConnectorId: aws.String(id),
	}

	resp, err := conn.GetVoiceConnectorWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, chimesdkvoice.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if resp == nil || resp.VoiceConnector == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	if err != nil {
		return nil, err
	}

	return resp.VoiceConnector, nil
}

func findVoiceConnectorWithRetry(ctx context.Context, conn *chimesdkvoice.ChimeSDKVoice, isNewResource bool, id string) (*chimesdkvoice.VoiceConnector, error) {
	var resp *chimesdkvoice.VoiceConnector
	err := tfresource.Retry(ctx, 1*time.Minute, func() *retry.RetryError {
		var err error
		resp, err = findVoiceConnectorByID(ctx, conn, id)
		if isNewResource && tfresource.NotFound(err) {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	}, tfresource.WithDelay(5*time.Second))

	return resp, err
}
