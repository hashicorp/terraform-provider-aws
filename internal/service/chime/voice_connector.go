// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chime

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/chimesdkvoice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/chimesdkvoice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_region": {
				Type:             schema.TypeString,
				ForceNew:         true,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.VoiceConnectorAwsRegion](),
			},
			names.AttrName: {
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
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	createInput := &chimesdkvoice.CreateVoiceConnectorInput{
		Name:              aws.String(d.Get(names.AttrName).(string)),
		RequireEncryption: aws.Bool(d.Get("require_encryption").(bool)),
		Tags:              getTagsIn(ctx),
	}

	if v, ok := d.GetOk("aws_region"); ok {
		createInput.AwsRegion = awstypes.VoiceConnectorAwsRegion(v.(string))
	}

	resp, err := conn.CreateVoiceConnector(ctx, createInput)
	if err != nil || resp.VoiceConnector == nil {
		return sdkdiag.AppendErrorf(diags, "creating Chime Voice connector: %s", err)
	}

	d.SetId(aws.ToString(resp.VoiceConnector.VoiceConnectorId))

	return append(diags, resourceVoiceConnectorRead(ctx, d, meta)...)
}

func resourceVoiceConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	resp, err := FindVoiceConnectorResourceWithRetry(ctx, d.IsNewResource(), func() (*awstypes.VoiceConnector, error) {
		return findVoiceConnectorByID(ctx, conn, d.Id())
	})

	if tfresource.TimedOut(err) {
		resp, err = findVoiceConnectorByID(ctx, conn, d.Id())
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Chime Voice connector %s not found", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Voice Connector (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, resp.VoiceConnectorArn)
	d.Set("aws_region", resp.AwsRegion)
	d.Set("outbound_host_name", resp.OutboundHostName)
	d.Set("require_encryption", resp.RequireEncryption)
	d.Set(names.AttrName, resp.Name)

	return diags
}

func resourceVoiceConnectorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	if d.HasChanges(names.AttrName, "require_encryption") {
		updateInput := &chimesdkvoice.UpdateVoiceConnectorInput{
			VoiceConnectorId:  aws.String(d.Id()),
			Name:              aws.String(d.Get(names.AttrName).(string)),
			RequireEncryption: aws.Bool(d.Get("require_encryption").(bool)),
		}

		if _, err := conn.UpdateVoiceConnector(ctx, updateInput); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Voice connector (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceVoiceConnectorRead(ctx, d, meta)...)
}

func resourceVoiceConnectorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ChimeSDKVoiceClient(ctx)

	input := &chimesdkvoice.DeleteVoiceConnectorInput{
		VoiceConnectorId: aws.String(d.Id()),
	}

	if _, err := conn.DeleteVoiceConnector(ctx, input); err != nil {
		if errs.IsA[*awstypes.NotFoundException](err) {
			log.Printf("[WARN] Chime Voice connector %s not found", d.Id())
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Voice connector (%s)", d.Id())
	}

	return diags
}

func findVoiceConnectorByID(ctx context.Context, conn *chimesdkvoice.Client, id string) (*awstypes.VoiceConnector, error) {
	in := &chimesdkvoice.GetVoiceConnectorInput{
		VoiceConnectorId: aws.String(id),
	}

	resp, err := conn.GetVoiceConnector(ctx, in)

	if errs.IsA[*awstypes.NotFoundException](err) {
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
